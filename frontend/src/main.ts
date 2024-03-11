import { createApp } from "vue";
import "./style.css";
import Chat from "./pages/Chat.vue";
import Index from "./pages/Index.vue";
import App from "./App.vue";
import { createRouter, createWebHistory } from "vue-router";

const routes = [
  { path: "/", name: "index", component: Index },
  { path: "/r/:roomId", name: "room", component: Chat },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

const app = createApp(App);

app.use(router);
app.mount("#app");
