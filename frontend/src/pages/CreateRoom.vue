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
  <div class="mx-auto max-w-xs">
    <main v-if="!loading">
      <h1 class="my-8 text-2xl text-center">Welcome to the chat!</h1>
      <form @submit.prevent="createRoom" class="space-y-2">
        <button type="submit" class="btn btn-primary btn-block">
          Create new room
        </button>
      </form>
    </main>

    <div v-else class="flex justify-center items-center h-screen">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  </div>

  <div v-if="errorOccured" class="toast toast-center toast-middle">
    <div class="alert alert-error">
      <span>Unexpected error occured. Please, try later</span>
    </div>
  </div>
</template>
