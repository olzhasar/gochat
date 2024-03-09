<script setup lang="ts">
import { Ref, ref } from "vue";

interface Message {
    content: string;
    author: string | null;
}

let messages: Ref<Message[]> = ref([]);
let messagePrompt = ref("");
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
  if (messagePrompt.value === "" || messagePrompt.value === null) {
    return;
  }

  ws.send(messagePrompt.value);
  messages.value.push({ content: messagePrompt.value, author: null });
  messagePrompt.value = "";
  scrollToBottom();
};

const setName = () => {
  ws.send(name.value);
  nameSet.value = true;
  scrollToBottom();

  const element = document.getElementById("messageInput") as HTMLInputElement;
  window.setTimeout(() => element.focus(), 0);
};

</script>

<template>
  <div class="flex overflow-hidden flex-col mx-auto max-w-md h-screen">
    <h1 class="my-4 text-2xl text-center">Chat</h1>

    <form v-if="!nameSet" class="space-y-2" @submit="(event) => {event.preventDefault(); setName()}">
      <input tabindex="0" v-model="name" class="w-full input input-bordered" type="text" placeholder="Enter your name" autofocus />
      <button class="btn btn-primary btn-block">Start chatting</button>
    </form>

    <div v-show="nameSet" id="messageList" class="overflow-y-scroll flex-grow pr-2 my-4 space-y-4 no-scrollbar">
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
    </form>

  </div>
</template>
