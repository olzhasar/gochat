<script setup lang="ts">
const share = async () => {
  if ("share" in navigator) {
    try {
      await navigator.share({
        title: "Join my super secret chat room",
        text: window.location.href,
        url: window.location.href,
      });
    } catch (error) {
      console.error("Error sharing:", error);
    }
  } else {
    window.navigator.clipboard.writeText(window.location.href);
    alert(
      "Link copied to clipboard. Share it with others you want to join the chat room.",
    );
  }
};

const showModal = () => {
  const modal = document.getElementById("modal-home") as HTMLDialogElement;
  modal.showModal();
};
</script>

<template>
  <dialog id="modal-home" class="modal">
    <div class="modal-box">
      <h3 class="text-lg font-bold">Are you sure want to leave this room?</h3>
      <div class="modal-action">
        <form method="dialog">
          <!-- if there is a button in form, it will close the modal -->
          <a class="mr-4 btn btn-outline btn-primary" href="/">Yes</a>
          <button class="btn">No</button>
        </form>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button>close</button>
    </form>
  </dialog>

  <div class="flex navbar flex-0">
    <div class="flex-0">
      <button class="btn btn-square btn-ghost" @click="showModal">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          class="w-6 h-6"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="m2.25 12 8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25"
          />
        </svg>
      </button>
    </div>
    <div class="flex-grow text-center">
      <h1 class="w-full text-lg font-bold text-center">Super secret chat</h1>
    </div>
    <div class="flex-none">
      <button class="btn btn-square btn-ghost" @click="share">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          class="w-6 h-6"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0 0a2.25 2.25 0 1 0 3.935 2.186 2.25 2.25 0 0 0-3.935-2.186Zm0-12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z"
          />
        </svg>
      </button>
    </div>
  </div>
</template>
