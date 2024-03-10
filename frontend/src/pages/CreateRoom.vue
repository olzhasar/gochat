<script setup lang="ts">
import { useRouter } from "vue-router";

const router = useRouter();

const createRoom = () => {
  const apiUrl = import.meta.env.VITE_API_URL as string;
  const url = apiUrl + "/room";

  fetch(url, {
    method: "POST",
  })
    .then((response) => response.text())
    .then((data) => {
      console.log(data);
      router.push({ name: "room", params: { roomId: data } });
    });
};
</script>

<template>
  <div class="mx-auto max-w-xs">
    <h1 class="my-8 text-2xl text-center">Welcome to the chat!</h1>
    <form @submit.prevent="createRoom" class="space-y-2">
      <button type="submit" class="btn btn-primary btn-block">
        Create new room
      </button>
    </form>
  </div>
</template>
