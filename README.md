Start the server through powershell by cd'ing to the server folder and using the command: "go run server.go".
Optionally you can use "-port" and "-name" to set a given servername and server port. By default these will be "localhost" and "8080".
The server will output all actions and lamport timestamps in the log.server file inside the server folder.

Connect clients by opening a new powershell and cd to the client folder, from there run the command "go run client.go"
The program will then ask you to input the servername and port in the format "servername:port". If you wish to connect with default values just press enter with nothing,
entered. The default values will then be "localhost" and "8080".

A client can exit the program by typing "exit" as a message, or by pressing CTRL+C to terminate powershell application. The server will be aware of both of these shutdowns.
