Create Table for storing convos and participants in that convo
```
CREATE TABLE convos(
    convo_id SERIAL PRIMARY KEY,
    participants integer[]
);
```

Create Table for storing messages
```
CREATE TABLE messages(
    id SERIAL PRIMARY KEY,
    user_id int references users(user_id),
    convo_id int references convos(convo_id),
    content TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

Create Table for storing users
```
CREATE TABLE "users" (
	"user_id" SERIAL,
	"username" TEXT NOT NULL UNIQUE,
	"passwd_hash" TEXT NOT NULL,
	"created_at" TIMESTAMP,
	"admin" BOOLEAN NOT NULL DEFAULT '0',
	PRIMARY KEY ("user_id")
);
```

