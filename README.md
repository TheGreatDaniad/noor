# noor

Women, Life, Freedom 

Noor is an obfuscation-oriented tunnel designed to be used against internet censorship in Iran and other countries with restricted internet connection. 

Design goals: 

Noor is focused on obfuscation. Its purpose is to be working in the hardest situation where most of the other protocols does not work. There are also security and performance in the second place. It is meant to be flexible in case of updates in the cencorship infrastructure (a.k.a gfw) and supporting different approaches to do each of the internal functionalities so it would be resiliant in terms of different internet blocks of different regions, times, ISPs. Another design principle is simplicity and ease of use for both server providers and clients. For the clients it should work by just installing an app and clicking connect button and potentially a server address and a pass phrase with user id, or just simply scan a qr code. Also for the admins, it should be installed in a few minutes and least possible config, potentially zero config. 
Noor has two versions, one for private use and the other for public use. Public use is the network of free servers that servers public clients with no password. Private use is when the server adds clients manually and only certain people has access to it. 
Noor also values quality of service. For public servers there is some means to control server usage and prevent lack of quality of service due to overloading the servers. Number of connections in each server is restricted based on server resources, and every connection has a cap of bandwidth calculated based on real-time usage stats of the whole system. There possibly should be a queue for the users to connect the service to prevent service overload. 

Cryptography: 

This part should satisfy both security and obfuscation. As for now, my decision is to use AES-256 cryptography algorithm to encrypt packets. The 32 bit secret phrase is derived from hashing (SHA-256) the user password. Hence, each user of a server is identified by its password, and passwords must be unique. Each user can also have other data such as name or email in the server for the sake of statistics, but the client only needs the server address and the password in order to be able to connect. 

Handshaking:

Since the handshaking is where most of the protocols get blocked by the gfw, there should be smart and varius way to handshake. Mechanisms for handshaking must be quite flexable and adabtable to the updates of the gfw. 
As for now, the handshaking is done using http request containing a json web token (jwt). Token contains user id and is signed and encrypted using the user password. 

