import websockets
import requests
import asyncio
import threading
import json
import urwid

class Client:
    def __init__(self, username, password, receiver, callback=None):
        self.login(username, password)
        self.username = username
        self.receiver = receiver
        self.callback = callback

    def login(self, username, password):
        response = requests.post('http://localhost:8080/login', json={'username': username, 'passwd_hash': password})
        if response.status_code != 200:
            raise Exception('POST /login/ {}'.format(response.status_code))
        self.token = response.json()['token']

    async def read_message(self, websocket):
        while True:
            try:
                message = await websocket.recv()
                message = json.loads(message)
                if self.callback:
                    self.callback(message)

            except websockets.ConnectionClosed:
                print("Connection closed. Reconnecting...")
                break

    async def handle_user_input(self, websocket):
        while True:
            message = input("Enter message: ")
            await self.send_message(websocket, message)

    async def connect_and_handle_messages(self):
        uri = "ws://localhost:8080/upgrade/{}"
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.token}",
        }
        while True:
            try:
                async with websockets.connect(uri.format(self.receiver), extra_headers=headers) as websocket:
                    self.websocket = websocket
                    await self.read_message(websocket)
            except Exception as e:
                print(f"Error during connection: {e}")
                await asyncio.sleep(500)  # Pause before retrying

    async def send_message(self, message):
        ws = self.websocket
        data = {
            "type": "message",
            "sender": self.username,
            "recipient": self.receiver,
            "content": message,
        }
        data_bytes = json.dumps(data).encode('utf-8')
        await ws.send(data_bytes)


    async def run(self, lw):
        self.lw = lw
        await self.connect_and_handle_messages()
        
if __name__ == "__main__":
    username = "check2"
    username = input("Enter username: ")
    password = "test"
    receiver = "check3"
    client = Client(username, password, receiver)
    asyncio.run(client.run())


