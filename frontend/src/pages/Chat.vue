<script setup lang="ts">
import { useRoute } from "vue-router";

const route = useRoute();
const roomId = route.params.roomId;

const apiURL = import.meta.env.VITE_API_URL as string;
const wsURL = apiURL.replace("http", "ws");

const roomURL = `${wsURL}/room/${roomId}`;

import { Ref, onMounted, ref } from "vue";

enum MessageType {
  TEXT = 1,
  NAME = 2,
  LEAVE = 3,
  TYPING = 4,
  STOP_TYPING = 5,
}

interface Message {
  msgType: number;
  content: string;
  author: string | null;
}

let connected = ref(false);
let messages: Ref<Message[]> = ref([]);
let messagePrompt = ref("");
let name = ref("");
let nameSet = ref(false);
let typing = ref(false);
let personTyping = ref("");

const scrollToBottom = () => {
  const messagesDiv = document.getElementById("messageList") as HTMLElement;
  setTimeout(() => (messagesDiv.scrollTop = messagesDiv.scrollHeight), 0);
};

let ws: WebSocket;

const connect = (): WebSocket => {
  ws = new WebSocket(roomURL);

  ws.onopen = () => {
    console.log("connected");
    connected.value = true;

    if (nameSet.value) {
      console.log("Sending name: " + name.value);
      ws.send(name.value);
    }
  };
  ws.onmessage = (event) => {
    receiveMessage(event);
  };
  ws.onclose = () => {
    console.log("disconnected");
    connected.value = false;
    setTimeout(() => {
      ws = connect();
    }, 2000);
  };
  ws.onerror = () => {
    console.log("error encountered. closing");
    ws.close();
  };

  return ws;
};

onMounted(() => {
  ws = connect();
});

const parseMessage = (rawMessage: string): Message => {
  const msgType = parseInt(rawMessage.charAt(0));
  const [author, content] = rawMessage.slice(1).split("|");
  return { msgType: msgType, content: content, author: author };
};

const receiveMessage = (event: MessageEvent) => {
  const msg = parseMessage(event.data);

  if (msg.msgType === MessageType.TYPING) {
    const author = msg.author as string;
    personTyping.value = author;
    scrollToBottom();
    return;
  }

  if (
    msg.msgType === MessageType.STOP_TYPING &&
    personTyping.value === msg.author
  ) {
    console.log("stopped typing: " + msg.author);
    personTyping.value = "";
    return;
  }

  if (msg.msgType === MessageType.LEAVE && msg.author === personTyping.value) {
    personTyping.value = "";
  }

  if (msg.author !== null && msg.author === personTyping.value) {
    personTyping.value = "";
  }

  messages.value.push(msg);
  scrollToBottom();
};

const sendToWS = (message: string, messageType: MessageType) => {
  const content = `${messageType}${message}`;
  ws.send(content);
};

const sendMessage = () => {
  if (messagePrompt.value === "" || messagePrompt.value === null) {
    return;
  }

  sendToWS(messagePrompt.value, MessageType.TEXT);
  messages.value.push({
    msgType: MessageType.TEXT,
    content: messagePrompt.value,
    author: null,
  });
  messagePrompt.value = "";
  typing.value = false;
  scrollToBottom();
  focusMessageInput();
};

const setName = () => {
  sendToWS(name.value, MessageType.NAME);
  nameSet.value = true;
  scrollToBottom();

  focusMessageInput();
};

const focusMessageInput = () => {
  const element = document.getElementById("messageInput") as HTMLInputElement;
  setTimeout(() => element.focus(), 0);
};

const notifyTyping = () => {
  if (!typing.value) {
    typing.value = true;
    sendToWS("", MessageType.TYPING);
  }
};

const notifyStopTyping = () => {
  if (typing.value) {
    typing.value = false;
    sendToWS("", MessageType.STOP_TYPING);
  }
};
</script>

<template>
  <div
    class="flex overflow-hidden flex-col p-2 mx-auto max-w-md h-screen md:p-0"
  >
    <main v-if="connected">
      <h1 class="my-4 text-2xl text-center">Chat</h1>

      <form
        v-if="!nameSet"
        class="space-y-2"
        @submit="
          (event) => {
            event.preventDefault();
            setName();
          }
        "
      >
        <input
          tabindex="0"
          v-model="name"
          class="w-full input input-bordered"
          type="text"
          placeholder="Enter your name"
          autofocus
        />
        <button class="btn btn-primary btn-block">Start chatting</button>
      </form>

      <div
        v-show="nameSet"
        id="messageList"
        class="overflow-y-scroll flex-grow pr-2 my-4 space-y-2 no-scrollbar"
      >
        <div v-for="msg in messages">
          <div
            v-if="msg.msgType == MessageType.TEXT && msg.author != null"
            class="chat chat-start"
          >
            <div class="chat-header">{{ msg.author }}</div>
            <div class="chat-bubble chat-bubble-primary">
              {{ msg.content }}
            </div>
          </div>

          <div
            v-if="msg.msgType == MessageType.TEXT && msg.author === null"
            class="chat chat-end"
          >
            <div class="chat-bubble">{{ msg.content }}</div>
          </div>

          <p v-if="msg.msgType == MessageType.NAME" class="my-2 text-center">
            {{ msg.author }} has joined the chat
          </p>

          <p v-if="msg.msgType == MessageType.LEAVE" class="my-2 text-center">
            {{ msg.author }} has disconnected
          </p>
        </div>

        <div v-if="personTyping" class="chat chat-start">
          <div class="opacity-60 chat-footer">
            {{ personTyping }} is typing...
          </div>
        </div>
      </div>

      <form
        @submit="
          (event) => {
            event.preventDefault();
            sendMessage();
          }
        "
        class="flex gap-2 pr-0 my-4"
        id="messageForm"
        v-show="nameSet"
      >
        <input
          tabindex="1"
          id="messageInput"
          v-model="messagePrompt"
          class="w-full input input-bordered"
          type="text"
          placeholder="Type a message"
          @focus="scrollToBottom"
          @input="notifyTyping"
          @blur="notifyStopTyping"
        />

        <button class="btn btn-square btn-secondary" type="submit">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            stroke-width="1.5"
            stroke="currentColor"
            class="w-6 h-6"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5"
            />
          </svg>
        </button>
      </form>
    </main>

    <div v-else class="flex justify-center items-center h-full">
      <div class="flex flex-col items-center">
        <h1 class="text-2xl">Connecting...</h1>
        <span class="loading loading-spinner loading-lg"></span>
      </div>
    </div>
  </div>
</template>
