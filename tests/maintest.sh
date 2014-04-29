#!/bin/bash

if [ -z $GOPATH ]; then
    echo "FAIL: GOPATH environment variable is not set"
    exit 1
fi

# Build the srunner binary to use to test the student's storage server implementation.
# Exit immediately if there was a compile-time error.
go install github.com/cmu440/goplaysgo/runners/mrunner
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

go install github.com/cmu440/goplaysgo/tests/maintest
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi


# Pick random port between [10000, 20000).
MASTER_PORT=$(((RANDOM % 10000) + 10000))
MAIN_SERVER=$GOPATH/bin/mrunner
MAIN_TEST=$GOPATH/bin/maintest

function startMainServers {
    N=${#STORAGE_ID[@]}
    # Start master storage server.
    ${MAIN_SERVER} -N=${N} -port=${MASTER_PORT} 2> /dev/null &
    MAIN_SERVER_PID[0]=$!
    # Start slave storage servers.
    if [ "$N" -gt 1 ]
    then
        for i in `seq 1 $((N-1))`
        do
	          MAIN_SLAVE_PORT=$(((RANDOM % 10000) + 10000))
            ${MAIN_SERVER} -port=${MAIN_SLAVE_PORT} -master="localhost:${MASTER_PORT}" 2> /dev/null &
            MAIN_SERVER_PID[$i]=$!
        done
    fi
    sleep 5
}

function stopMainServers {
    N=${#STORAGE_ID[@]}
    for i in `seq 0 $((N-1))`
    do
        kill -9 ${MAIN_SERVER_PID[$i]}
        wait ${MAIN_SERVER_PID[$i]} 2> /dev/null
    done
}

function testRouting {
    startMainServers
    for KEY in "${KEYS[@]}"
    do
        ${MAIN_TEST} -port=${STORAGE_PORT} p ${KEY} value > /dev/null
        PASS=`${LRUNNER} -port=${STORAGE_PORT} g ${KEY} | grep value | wc -l`
        if [ "$PASS" -ne 1 ]
        then
            break
        fi
    done
    if [ "$PASS" -eq 1 ]
    then
        echo "PASS"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo "FAIL"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
    stopStorageServers
}

# Testing Single Server
function testSingle {
  echo "Running Single:"
  STORAGE_ID=('1')
  startMainServers
  ${MAIN_TEST} -port="localhost:${MASTER_PORT}"
  stopMainServers
}

function testTriple {
  echo "Running Triple:"
  STORAGE_ID=('1', '2', '3')
  startMainServers
  ${MAIN_TEST} -port="localhost:${MASTER_PORT}"
  stopMainServers
}

function testMany {
  echo "Running Many:"
  STORAGE_ID=('1', '2', '3', '4', '5')
  startMainServers
  ${MAIN_TEST} -port="localhost:${MASTER_PORT}"
  stopMainServers
}

# Run tests.
#testSingle
testTriple
testMany
