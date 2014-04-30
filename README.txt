                                               
 _____        _____ _                _____     
|   __|___   |  _  | |___ _ _ ___   |   __|___ 
|  |  | . |  |   __| | .'| | |_ -|  |  |  | . |
|_____|___|  |__|  |_|__,|_  |___|  |_____|___|
                         |___|                 
--------------------------------------------
Description
--------------------------------------------
Go Plays Go is a server where AI code for the game Go can be submitted to play
against other AI code. The system has three main components. First, is the 
MainServer, which is the place where AI code is submitted and the game results
are stored. Second, there are AI servers. The AI servers are where the 
submitted AI code runs and keeps track of game sttaes for individual games.
Finally, we have the web server, where users can code up their own go AI, and 
submit to the Main Servers.

--------------------------------------------
Runners
--------------------------------------------
Main Server Runner:        bin/mrunner  (from github.com/cmu440/goplaysgo/runners/mrunner)
AI Server Runner:          bin/airunner (from github.com/cmu440/goplaysgo/runners/airunner)
Main Server Client Runner: bin/crunner  (from github.com/cmu440/goplaysgo/runners/crunner)

--------------------------------------------
Tests
--------------------------------------------
The test scripts are setup so that they run the tests in the following order:
1) Single MainServer
2) Three MainServers
3) Five MainServers
4) Five MainServers with 2 of them Killed

Main Server Tests:  
  tests/maintest.sh (github.com/cmu440/goplaysgo/tests/maintest)

The Main Server Tests are tests that check the basuc operations of the Main
Servers. There are 5 tests included:
1) TestNormalSingle: 
   Submits AIs to a single Main Server and checks correctness
2) TestNormalMultiple: 
   Submits AIs to multiple Main Servers and checks correctness
3) TestCompileError:
   Test that Main Server Rejects Compile Errors
4) TestDuplicateSingle:
   Tests if duplicate AI (AI with the same name) is rejected with single server
5) TestDuplicateMultiple:
   Tests if duplicate AI submitted in multiple servers is rejected

Paxos Tests:
  tests/paxostest.sh (github.com/cmu440/goplaysgo/tests/paxostest)

The Paxos Tests test the funtionality of the Paxos RPC calls.
Three are 4 tests included:
1) TestNormal:
   Tests the normal Paxos Protocol, where a value is successfully commited
2) TestPrepare:
   Tests the Prepare phase of paxos, where it should reject certain requests
3) TestAccept:
   Tests the Accept phase of paxos, where it should reject certain requests
4) TestCatchUp:
   Tests the functionality of the servers catching up where some nodes can be 
   behind/missing commands. Commits first command to 2f+1 nodes and the 
   second command to the other 2f+1 nodes. Then it commits a Nop command (to 
   get the nodes synced up), and checks that all three are on the same state


-------------------------------------
Design Decisions
-------------------------------------
TODO
