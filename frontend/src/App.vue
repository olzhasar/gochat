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
  scrollToBottom();
};
ws.onclose = () => {
  console.log("disconnected");
};

const sendMessage = () => {
  if (messageInput.value === "" || messageInput.value === null) {
    return;
  }

  ws.send(messageInput.value);
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
  <div class="flex overflow-hidden flex-col mx-auto max-w-md h-screen">
    <h1 class="my-4 text-2xl text-center">Chat</h1>

    <form v-if="!nameSet" class="space-y-2" @submit="(event) => {event.preventDefault(); setName()}">
      <input v-model="name" class="w-full input input-bordered" type="text" placeholder="Enter your name" />
      <button class="btn btn-primary btn-block">Start chatting</button>
    </form>

    <div v-show="nameSet" id="messageList" class="overflow-y-scroll flex-grow pr-2 my-4 space-y-4">
      <div v-for="msg in messages">
	<div v-if="msg.author != null" class="chat chat-start">
	  <div class="chat-header">{{ msg.author }}</div>
	  <div class="chat-bubble chat-bubble-primary">
	    {{ msg.content }}
	  </div>
	</div>

	<div v-else class="chat chat-end">
	  <div class="chat-bubble chat-bubble-secondary">{{ msg.content }}</div>
	</div>

      </div>
    </div>

    <form @submit="(event) => {
      event.preventDefault();
      sendMessage();
    }
      "
      class="my-4"
    >
      <input
	v-model="messageInput"
	v-if="nameSet"
	class="w-full input input-bordered"
	type="text"
	placeholder="Type a message"
      />
    </form>

  </div>
</template>
