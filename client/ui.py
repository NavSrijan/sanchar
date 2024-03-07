import urwid
from client import Client
import threading
import asyncio


text_box = urwid.Edit("> ")
lw = urwid.SimpleListWalker([])
listbox = urwid.AttrMap(
    urwid.ListBox(lw),
    "listbox",
)
output_box = urwid.Frame(listbox)
frame_widget = urwid.Frame(
        body=output_box,
        footer=text_box,
        focus_part='footer')

def exit_on_enter(key: str | tuple[str, int, int, int]) -> None:
    if key == "enter":
        raise urwid.ExitMainLoop()


def run_client():
    asyncio.run(client.run(lw))

def send_message(data):
    asyncio.run(client.send_message(data))

def check_for_command(data):
    if not data or data[0] != "/" or len(data) == 1 or data[1] == "/":
        return False
    return True

def parse_command(data):
    command = data[1:]
    if command == "exit":
        raise urwid.ExitMainLoop()

def on_send_message(data):
    message = text_box.get_edit_text()
    text_box.set_edit_text("")
    if check_for_command(message):
        parse_command(message)
    else:
        send_message(message)

def on_receive_message(message):
    lw.append(urwid.Text(f">{message['content']}\n"))
    listbox.original_widget.set_focus(len(lw) - 1, "above")
    loop.draw_screen()


username = "check2"
username = input("Enter username: ")
password = "test"
receiver = "check3"
receiver = input("Enter receiver: ")
client = Client(username, password, receiver, on_receive_message)
thread = threading.Thread(target=run_client)
thread.start()

loop = urwid.MainLoop(frame_widget, unhandled_input=on_send_message)
loop.run()


