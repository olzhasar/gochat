<script setup lang="ts">
const ws = new WebSocket("ws://localhost:8080/ws");
ws.onopen = () => {
  console.log("connected");
};
ws.onmessage = (event) => {
  console.log(event.data);
};
ws.onclose = () => {
  console.log("disconnected");
};

let message = "";

const sendMessage = () => {
  ws.send(message);
  message = "";
};
</script>

<template>
  <div class="text-center">
    <h1 class="my-4 text-2xl">Chat</h1>

    <input
      v-model="message"
      @keyup.enter="sendMessage"
      class="p-2 w-1/2 border-2 border-gray-300"
      type="text"
      placeholder="Type a message"
    />
  </div>
</template>
