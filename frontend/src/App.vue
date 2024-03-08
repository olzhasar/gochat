<script setup lang="ts">
import { Ref, ref } from "vue";

let messages: Ref<string[]> = ref([]);
let message = ref("");

const ws = new WebSocket("ws://localhost:8080/ws");
ws.onopen = () => {
  console.log("connected");
};
ws.onmessage = (event) => {
  const content = event.data as string;
  messages.value.push(content);
};
ws.onclose = () => {
  console.log("disconnected");
};

const sendMessage = () => {
  ws.send(message.value);
  console.log("sent: ", message.value);
  message.value = "";
};
</script>

<template>
  <div class="mx-auto max-w-md">
    <h1 class="my-4 text-2xl text-center">Chat</h1>

    <input
      v-model="message"
      @keyup.enter="sendMessage"
      class="p-2 w-full rounded-md border shadow"
      type="text"
      placeholder="Type a message"
    />

    <div class="my-4 space-y-4 text-slate-600">
      <div v-for="msg in messages" class="py-1">
        <span class="py-2 px-4 bg-white rounded-md border shadow">{{
          msg
        }}</span>
      </div>
    </div>
  </div>
</template>
