                                               
 _____        _____ _                _____     
|   __|___   |  _  | |___ _ _ ___   |   __|___ 
|  |  | . |  |   __| | .'| | |_ -|  |  |  | . |
|_____|___|  |__|  |_|__,|_  |___|  |_____|___|
                         |___|                 

By: Won Gu Kang (wonguk@andrew) - Main & AI Servers
    Dale Best   (dbest@andrew)  - Go Game Logic & Webserver

Please feel free to contact us if you have any questions about the system :)

----------------------------------------------
Description
----------------------------------------------
Go Plays Go is a server where AI code for the game Go can be submitted to play
against other AI code. The system has three main components. First, is the 
MainServer, which is the place where AI code is submitted and the game results
are stored. Second, there are AI servers. The AI servers are where the 
submitted AI code runs and keeps track of game sttaes for individual games.
Finally, we have the web server, where users can code up their own go AI, and 
submit to the Main Servers.

---------------------------------------------
Runners
---------------------------------------------
Main Server Runner:        bin/mrunner  
  Package: github.com/cmu440/goplaysgo/runners/mrunner
AI Server Runner:          bin/airunner 
  Package: github.com/cmu440/goplaysgo/runners/airunner
Main Server Client Runner: bin/crunner
  Package: github.com/cmu440/goplaysgo/runners/crunner

--------------------------------------------
Sample Executions
--------------------------------------------
# Export GOPATH
export GOPATH=

# Start Main Servers
bin/mrunner -port=12300 -N=3
bin/mrunner -port=12301 -N=3 -master=localhost:12300
bin/mrunner -port=12302 -N=3 -master=localhost:12300

# Submit AIs
bin/crunner -port=12300 sa test0 ai_examples/ai_random.go
bin/crunner -port=12300 sa test1 ai_examples/ai_random.go
bin/crunner -port=12301 sa test2 ai_examples/ai_random.go

# Check Standings
bin/crunner -port=12302 sd

# Kill Server w/ port 12302 and check Server still works
# Note: To see the standings, it may take longer because the new AI server
#       tries to connect to the dead Main Server
# Note: When you ctrl-c from the mrunner, it seems to kill the AI servers
#       that the specific main server started. So please don't panic if 
#       you don't see some game results between different AIs. 
#       On the other hand, The test script also kills the mrunners, but 
#       the ai servers there don't seem to be killed as well.... so the 
#       tests should verify that the correct behavior.
bin/crunner -port=12301 sa test3 ai_examples/ai_random.go
bin/crunner -port=12300 sd

# Start New Replacement Server
bin/mrunner -port=12303 -r

# Manually Add new Server to Paxos Ring
# Setup Quiese Mode to existing servers
bin/crunner -port=12300 qs
bin/crunner -port=12301 qs

# Sync up the servers to a given command number
# Note: This step is not necessary if the servers are already at the same
#       command number.
bin/crunner -port=12300 -cmdNum=20 qs
bin/crunner -port=12301 -cmdNum=20 qs

# Replace the dead server with new server and bring the new server upto speed
bin/crunner -port=12300 -add=localhost:12303 -replace=12302 -master qr
# Replace the dead server with the replacement in existing server
bin/crunner -port=12300 -add=localhost:12303 -replace=12302 qr

# Check new server is updated
bin/crunner -port=12303 sd

# Note: If there are less than the majority of servers remaining,
#       a request to submit a new AI will basically hang forever, because
#       the AI will never be commited into the system. I wasn't sure
#       how that could be tested, so you should try it out :)

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
WebServer/Front End
-------------------------------------
The file webserver.go contains code that initiates the user front end website. In order to run this please compile webserver.go then run the application. While it is running in a local browser go to localhost http://localhost:8080/edit/(user) in which (user) is replaced by a username of your choice.

On the first page (edit page) the user can write up go coe for the A.I which should be in the format of:
package ai

import (
  "github.com/cmu440/goplaysgo/gogame"
  ...
  )

func NextMove(board gogame.Board, player gogame.Player) gogame.Move {
...
}

For more help the user can check example code in the ai_example folder
and look at the ai/ai_api.go file.


-------------------------------------
Design Decisions
-------------------------------------
Main Server
  The Main Server has 2 sets of RPC APIs:
  - MainServer API (mainserver/mainserver_api.go)
    This API is used by external clients/servers to communicate to the
    main server.

  - Paxos API (mainserver/paxos_api.go)
    This API is used by the Main servers to communicate with each other
    through the Paxos Protocols. It also included the RPC call necessary
    to manually replace dead servers.

  The Main Server consists of 3 "Masters":
  - AI Master (mainserver/ai_master.go)
    The AI Master is responsible for compiling and launching the AI Servers
    for the AIs submitted by the clients. The AI Master also "schedules" the
    AI Servers by sending them all the AIs that are already in the system, so
    that all AI Servers play each other. There is no fuctionality to restart
    dead AI Servers. The main reason behind this is because that probably 
    indicates that the AI code submitted raised an exception, which crashed
    the AI Server, which could be interpretted as the AI not being as good as 
    the other AIs.

  - Stats Master (mainserver/stats_master.go)
    The Stats Master is responsible for keeping track of all the commited AIs
    and game results in the Main servers.

  - Paxos Master (mainserver/paxos_master.go)
    The Paxos Master uses the Paxos protocol to commit command requests from
    the Stats Master and game results from the AI servers into the paxos group.
    The Paxos Master also makes sure that only commands

  Basic Flow:
  - Submitting an AI:
    1) The SubmitAI RPC call sends the AI info to the AI Master
    2) The AI Master checks if it has seen the AI before
    3) The AI Master compiles and launches the AI Server. Returns CompileError
       if the code fails to compile.
    4) The AI Master Sends the AI Info to the Stats Master to commit into the
       system.
    5) The Stats Master passes AI Info to Paxos Master to commit into Paxos ring
    6) The Stats Master tells the AI Master whether or not the AI is a duplicate
    7) AI Master Sends all the existing AIs to the AI Server
    8) AI Server plays game against all the AIs and returns each game result
       back to Main Server

AI Server
  The AI Server has a single RPC API:
  - The AIServer API:
    This API is used by AI Servers to play games with each other, and by main
    servers to send the AIs the AI server should play against
  
  The AI Server has single Master:
  - The Game Master:
    The Game Master bsaically keeps track of all the games the AI is playing.
    When given the list of AIs to play against, the Game Master will do a
    2PC like protocol where it checks that itself and the opponent can play the
    game, and then initializes the game on both servers, and then plays the 
    game. We decided that we will let the Game Masters be referees themselves
    by using the 'gogame' package. This means that for a given game, both AIs
    keep track of the game and tell each other which moves they made.
    We also deicded that the server who starts the game should report back to
    the main server.

The Main reason we separated the AI Server and the Main server was because we 
wanted to modularize the system more. We initially thought of running the 
games in a thread on the Main Server, but it would have costed too much to 
replicate that through paxos, and also handle errors. Also, Go(golang) did not
support dynamically runing the AI code.

We also thought of having the AI Master be responsible for scheduling each
and individual game, but we found out that would not be atomic if the main 
server were to crash. As a result, we gave that responsibility to the ai 
servers to handle, and as mentioned above, if an AI Server crashes, we don't 
care because it is probably the AI code that is unstable and has caused 
the system to crash.

Finally, we decided to let the AI Servers make an RPC Call to the main servers
once a game has ended. Initially, we thought of making a single RPC call
from the main server to start a game and get the game results in the reply
object recieved when the RPC call had returned. This was a problem because if 
the main server that made the RPC call were to have died, the results would 
have been lost. So we decided to give the AI Servers a list of all the main
servers in the paxos ring to be able to send game results even if there were
partial failures in the main servers.

Again, feel free to send us an email if you have any questions about the system!
