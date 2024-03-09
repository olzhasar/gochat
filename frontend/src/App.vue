<script setup lang="ts">
const apiURL = import.meta.env.VITE_API_URL as string;

import { Ref, ref } from "vue";

interface Message {
    content: string;
    author: string | null;
}

let messages: Ref<Message[]> = ref([]);
let messagePrompt = ref("");
let name = ref("");
let nameSet = ref(false);

const ws = new WebSocket(apiURL);

const scrollToBottom = () => {
  const messagesDiv = document.getElementById("messageList") as HTMLElement;
  messagesDiv.scrollTop = messagesDiv.scrollHeight;
};

ws.onopen = () => {
  console.log("connected");
};
ws.onmessage = (event) => {
  const rawMessage = event.data as string;
  const [author, content] = rawMessage.split("|");
  messages.value.push({ content: content, author: author });
  window.setTimeout(scrollToBottom, 0);
};
ws.onclose = () => {
  console.log("disconnected");
};

const sendMessage = () => {
  if (messagePrompt.value === "" || messagePrompt.value === null) {
    return;
  }

  ws.send(messagePrompt.value);
  messages.value.push({ content: messagePrompt.value, author: null });
  messagePrompt.value = "";
  window.setTimeout(scrollToBottom, 0);
};

const setName = () => {
  ws.send(name.value);
  nameSet.value = true;
  window.setTimeout(scrollToBottom, 0);

  const element = document.getElementById("messageInput") as HTMLInputElement;
  window.setTimeout(() => element.focus(), 0);
};

</script>

<template>
  <div class="flex overflow-hidden flex-col p-2 mx-auto max-w-md h-screen md:p-0">
    <h1 class="my-4 text-2xl text-center">Chat</h1>

    <form v-if="!nameSet" class="space-y-2" @submit="(event) => {event.preventDefault(); setName()}">
      <input tabindex="0" v-model="name" class="w-full input input-bordered" type="text" placeholder="Enter your name" autofocus />
      <button class="btn btn-primary btn-block">Start chatting</button>
    </form>

    <div v-show="nameSet" id="messageList" class="overflow-y-scroll flex-grow pr-2 my-4 space-y-2 no-scrollbar">
      <div v-for="msg in messages">
	<div v-if="msg.author != null" class="chat chat-start">
	  <div class="chat-header">{{ msg.author }}</div>
	  <div class="chat-bubble chat-bubble-primary">
	    {{ msg.content }}
	  </div>
	</div>

	<div v-else class="chat chat-end">
	  <div class="chat-bubble">{{ msg.content }}</div>
	</div>

      </div>
    </div>

    <form @submit="(event) => {
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
      />

      <button class="btn btn-square btn-secondary" type="submit">
	<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
  <path stroke-linecap="round" stroke-linejoin="round" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5" />
</svg>
      </button>
    </form>

  </div>
</template>
