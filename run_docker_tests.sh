#!/usr/bin/env bash

# test related settings
DARKGRAY='\033[1;30m'
RED='\033[0;31m'
LIGHTRED='\033[1;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
LIGHTPURPLE='\033[1;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'

SET='\033[0m'

PASSED=0
FAILED=0

pass() {
  echo
  echo -e "${GREEN}PASS${SET}: $1"
  PASSED=$((PASSED+1))
  echo
}

fail() {
  echo -e "${RED}FAIL${SET}: $1"
  FAILED=$((FAILED+1))
}

report_test_result() {
  echo "Total PASSED: $PASSED"
  echo "Total FAILED: $FAILED"
}
trap report_test_result EXIT

httpgo -port 8000 &

assertion="Can update health"
echo "Negative case: All should return 503 after POST health?value=false"

# for loop in case httpgo is not ready yet
ok=0
for i in {1..30}; do
  status=$(curl -s -o /dev/null -w '%{http_code}' -XPOST http://localhost:8000/health?value=false)
  if [ "$status" != "200" ]; then
    echo "status=$status, sleep 2 seconds then retry..."
    sleep 2s
    continue
  else
    ok=1
    break
  fi
done

if [ $ok -ne 1 ]; then
  fail "$assertion"
else
  pass "$assertion"
fi

urls=("debug" "echo/x" "delay/123" "status/200" "health")
for i in "${urls[@]}"; do
  status=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8000/${i})
  subAssertion="GET ${i} $status"
  if [ "$status" == "503" ]; then
    pass "$subAssertion"
  else
    fail "$subAssertion"
  fi
done
echo "Positive case: All should return 200 after POST health?value=true"
status=$(curl -s -o /dev/null -w '%{http_code}' -XPOST http://localhost:8000/health?value=true)
if [ "$status" != "200" ]; then
  fail "$assertion"
fi
urls=("debug" "echo/x" "delay/123" "status/200" "health")
for i in "${urls[@]}"; do
  status=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8000/${i})
  subAssertion="GET ${i} $status"
  if [ "$status" == "200" ]; then
    pass "$subAssertion"
  else
    fail "$subAssertion"
  fi
done

exit ${FAILED}