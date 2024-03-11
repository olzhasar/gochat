<script setup lang="ts">
import { useRouter } from "vue-router";
import { ref } from "vue";

const router = useRouter();

let loading = ref(false);
let errorOccured = ref(false);

const createRoom = () => {
  const apiUrl = import.meta.env.VITE_API_URL as string;
  const url = apiUrl + "/room";

  loading.value = true;

  fetch(url, {
    method: "POST",
  })
    .then((response) => response.text())
    .then((data) => {
      router.push({ name: "room", params: { roomId: data } });
    })
    .catch((error) => {
      console.error("API request failed:", error);
      errorOccured.value = true;
      loading.value = false;
    });
};
</script>

<template>
  <main class="min-h-screen hero bg-base-200">
    <div class="text-center hero-content">
      <div v-if="!loading" class="max-w-md">
        <h1 class="text-5xl font-bold">Let's chat!</h1>
        <div>
          <ul class="block my-8 steps steps-vertical">
            <li class="step step-primary">Create room</li>
            <li class="step step-primary">Invite friends by sharing the URL</li>
            <li class="step step-primary">Enjoy the conversation!</li>
          </ul>
        </div>
        <button @click="createRoom" class="btn btn-primary">Create room</button>

        <ul
          class="p-4 mt-8 max-w-sm text-sm text-left text-gray-500 rounded-lg bg-base-200"
        >
          <li>
            The server does not store any conversations. Hence, there is no
            history in chat rooms.
          </li>
          <li>
            All rooms disappear after 1 minute of inactivity when all members
            leave
          </li>
          <li>
            Real-time communication is powered by encrypted
            <a class="link" href="https://en.wikipedia.org/wiki/WebSocket#"
              >Websocket</a
            >
            connections
          </li>
        </ul>

        <div class="mt-4">
          Created by
          <a class="link link-secondary" href="https://github.com/olzhasar"
            >@olzhasar</a
          >
        </div>
      </div>
      <div v-else>
        <span class="loading loading-spinner loading-lg"></span>
      </div>
    </div>
  </main>

  <div v-if="errorOccured" class="toast toast-center toast-top">
    <div class="alert alert-error">
      <span>Unexpected error occured. Please, try later</span>
    </div>
  </div>
</template>
