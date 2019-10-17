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

assertion="health should return 200"
status=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8000/health)
if [ "$status" == "200" ]; then
  pass "$assertion"
else
  fail "$assertion"
fi

exit $FAILED