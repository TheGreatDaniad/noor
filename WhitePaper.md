# Noor: An Obfuscation-oriented VPN Tunnel (Private mode)

## Introduction


Noor is a tunnel designed to combat internet censorship in countries such as Iran where access to the internet is restricted. It focuses on obfuscation and aims to work in the most challenging environments where most other protocols fail. In addition to security and performance, the protocol values simplicity, flexibility, and ease of use for both server providers and clients. This white paper outlines the design goals, cryptography, and handshaking mechanisms of the Noor VPN protocol.
 Noor has two versions: one for private use and one for public use. Public use is the network of free servers that serves public clients with no password, while private use is when the server adds clients manually and only certain people have access to it.
 In this paper we analyze the private side of Noor. 

## Design Goals

Noor is designed to be flexible and resilient to updates in censorship infrastructure and support different approaches to its internal functionalities. The protocol values simplicity and ease of use for both server providers and clients, with a user-friendly installation process and potentially zero configuration for normal users while it gives customizability for more professional users.
The protocol also values quality of service, with means to control server usage and prevent lack of quality of service due to server overload.

### Flexibility of Design
Every aspect of a vpn protocol that might be used by cencorship systems to block the protocol must be flexible and easily changeable in order to provide true obfuscation. 
This incluedes cryptography algorythms used to encrypt and decrypt data, key exchange algorithms, handshaking methods, authentication methods, OSI network layer of operation and more. 
For each of the items above there must be different implementations to be used in different situation to guarantee the service functionality. 


### forward compatibility 
One of the most important aspect of this protocol is to provide forward compatibility. Since cencorship tools are being updated regularly, the means to oppose it also must be updated regularly. Hence, the design of the system must be in a way that welcomes new changes and accomodate potential future changes in its heart. All of the headers that are added to the payload data must assume that they can change in the future. The client softwares must be already compatible with future changes as much as possible in a way that the updates are mostly required by the server nodes and not for the clients. 



## Cryptography



Noor prioritizes both security and obfuscation in its cryptography.
The encryption is a MUST in noor in order too keep the transfered data safe from authority probs to make it harder to find patterns in the packets plus the security of user data. Encryption/Decryption of data is done using a symmetric algorithm. For the time of writing this paper the chosen algorithm is AES-256-cbc but it can have other options in the future. The key exchange is done using a Curve25519 method. 

## Authentication 

Each user have a randomly generated 16 bit unsigned integer ID alongside with a minimum 8 character password. This password is used for the sake of authentication and not the encryption of the data. Using this password and ID users can handshake with the server and obtain a symmetric key for the encryption of their data in the session. The data that each user needs to connect to a server in 1-Server Address 2-Server Port 3-User ID 4-User Password that can be used in a client using a connection file or just manually inserted. There is also an option for scanning this data via a qr code. 




## OSI Model Operation Layer

Noor has the capacity to operate across different network layers while being highly customizable. Specifically, it can operate at Layer 3 by manipulating and transmitting IP packets, and at Layer 4 using TCP packets to send IP packets via a TCP channel or UDP. This flexibility in operation offers several advantages, such as the ability to customize and optimize the VPN software for specific use cases and scenarios.

Working at Layer 3, the VPN software can provide secure communication by changing and transmitting IP packets through the network. In contrast, operating at Layer 4 provides the additional capability of utilizing TCP packets, which offers greater control over data transmission and can lead to improved reliability and stability of the connection. Furthermore, the use of UDP in Layer 4 can help improve performance and reduce latency, which can be critical for real-time communication applications.

While in TCP mode it transfers encrypted IP packets between client and the server, in UDP mode or Layer 3 mode it attaches some more headers to route packets and maintain integrity and confidentiality. 


## Handshaking Methods
the main approach for handshaking is that a client sends a init message to the server. The only important data in the init message is the user ID and optionally a handshake method id. The structure of the init message is explained later in this text. If the optional handshake method byte was specified and server supports that method, the server then sends a ack message and then both parties continue the process of handshaking based on the chosen method. If the server does not support that method it send a reject byte in the response and closes the connection, then the client must use another method and try again. If the init message does not contain a handshake method byte, the server will send a byte method in its ack response.

The init packet is as below form: 
first 2 bytes are the 16 bit user ID and the third byte is the handshake method byte. 
The handshake method takes value between 0xC0 (192) to  0xFF (255) the server sends back a list of supported handshake methods. If it is between 0x80 (128) to 0xBF (191) the server sends back its prefered method of handshake. If it is between 0x00 to 0x40 (64) the value will be encountered as handshake method number and server does not need to send methods by itself. 


### Simple Handshake Method 

