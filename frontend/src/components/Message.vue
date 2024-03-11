<script setup lang="ts">
import { PropType } from "vue";
import { MessageType, Message } from "@/types";

defineProps({
  msg: {
    type: Object as PropType<Message>,
    required: true,
  },
});
</script>

<template>
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
</template>
