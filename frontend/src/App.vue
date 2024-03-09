<script setup lang="ts">
import { Ref, ref } from "vue";

interface Message {
    content: string;
    author: string | null;
}

let messages: Ref<Message[]> = ref([]);
let messageInput = ref("");
let name = ref("");
let nameSet = ref(false);

const ws = new WebSocket("ws://localhost:8080/ws");

ws.onopen = () => {
  console.log("connected");
};
ws.onmessage = (event) => {
  const rawMessage = event.data as string;
  const [author, content] = rawMessage.split("|");
  messages.value.push({ content: content, author: author });
  scrollToBottom();
};
ws.onclose = () => {
  console.log("disconnected");
};

const scrollToBottom = () => {
  const messagesDiv = document.getElementById("messageList") as HTMLElement;
  messagesDiv.scrollTop = messagesDiv.scrollHeight;
};

const sendMessage = () => {
  if (messageInput.value === "" || messageInput.value === null) {
    return;
  }

  ws.send(messageInput.value);
  console.log("sent: ", messageInput.value);
  messages.value.push({ content: messageInput.value, author: null });
  messageInput.value = "";
  scrollToBottom();
};

const setName = () => {
  ws.send(name.value);
  nameSet.value = true;
};

</script>

<template>
  <div class="overflow-hidden mx-auto max-w-md h-screen">
    <h1 class="my-4 text-2xl text-center">Chat</h1>

    <form v-if="!nameSet" class="space-y-2" @submit="(event) => {event.preventDefault(); setName()}">
      <input v-model="name" class="p-2 w-full rounded-md border shadow" type="text" placeholder="Enter your name" />
      <button class="block py-2 px-4 w-full text-white bg-green-600 rounded-md border">Start chatting</button>
    </form>

    <div v-show="messages.length" id="messageList" class="overflow-y-scroll pr-2 my-4 space-y-4 text-slate-600 h-[calc(100vh-10rem)]">
      <div v-for="msg in messages" class="py-1">
	<div v-if="msg.author !== null" class="flex justify-start">
	  <span v-if="msg.author !== null" class="py-2 px-4 bg-white rounded-md border shadow">{{ msg.author }}: {{ msg.content }}</span>
	</div>
	<div v-else class="flex justify-end">
	  <span class="py-2 px-4 bg-blue-200 rounded-md border shadow">{{ msg.content }}</span>
	</div>
      </div>
    </div>

    <form @submit="(event) => {
      event.preventDefault();
      sendMessage();
    }
      ">
      <input
	v-model="messageInput"
	v-if="nameSet"
	class="p-2 w-full rounded-md border shadow"
	type="text"
	placeholder="Type a message"
      />
    </form>

  </div>
</template>
