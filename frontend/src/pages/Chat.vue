<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import Navbar from "../components/Navbar.vue";
import { MessageType, Message } from "@/types";
import MessageDiv from "@/components/Message.vue";

const route = useRoute();
const router = useRouter();
const roomId = route.params.roomId;

const apiURL = import.meta.env.VITE_API_URL as string;
const wsURL = apiURL.replace("http", "ws");

const roomURL = `${wsURL}/ws/${roomId}`;

import { Ref, onMounted, ref } from "vue";

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

const checkRoomExists = async (): Promise<boolean> => {
  const url = `${apiURL}/room/${roomId}`;

  try {
    const response = await fetch(url);
    return response.ok;
  } catch (error) {
    console.error("API request failed:", error);
    return false;
  }
};

let ws: WebSocket;

const connect = async () => {
  const roomExists = await checkRoomExists();

  if (!roomExists) {
    router.push({ name: "index" });
    throw new Error("Room not found");
  }
  ws = new WebSocket(roomURL);

  ws.onopen = () => {
    console.log("connected");
    connected.value = true;

    if (nameSet.value) {
      setName();
    }
  };
  ws.onmessage = (event) => {
    receiveMessage(event);
  };
  ws.onclose = () => {
    console.log("disconnected");
    connected.value = false;
    setTimeout(() => {
      connect();
    }, 2000);
  };
  ws.onerror = () => {
    console.log("error encountered. closing");
    ws.close();
  };
};

onMounted(() => {
  connect();
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
  <main
    v-if="connected"
    class="flex overflow-hidden flex-col px-2 pb-2 mx-auto max-w-md h-[calc(100dvh)] md:p-0"
  >
    <Navbar />
    <form v-if="!nameSet" class="space-y-2" @submit.prevent="setName">
      <input
        tabindex="0"
        v-model="name"
        class="w-full input input-bordered"
        type="text"
        placeholder="Enter your name"
        minlength="1"
        required
        autofocus
      />
      <button class="btn btn-primary btn-block">Start chatting</button>
    </form>

    <div
      v-show="nameSet"
      id="messageList"
      class="overflow-y-scroll flex-grow pr-2 space-y-2 no-scrollbar"
    >
      <TransitionGroup
        tag="div"
        enter-from-class="opacity-0 -translate-x-10"
        enter-to-class="opacity-100 translate-x-0"
        enter-active-class="duration-200 ease-in transform"
      >
        <div v-for="(msg, index) in messages" :key="index">
          <MessageDiv :msg="msg" />
        </div>
      </TransitionGroup>

      <div v-if="personTyping" class="chat chat-start">
        <div class="opacity-60 chat-footer">
          {{ personTyping }} is typing...
        </div>
      </div>
    </div>

    <form
      @submit.prevent="sendMessage"
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
        autocomplete="off"
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
</template>
