<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useNotificationsStore } from '@/stores/notifications'

const notificationsStore = useNotificationsStore()
const { notifications } = storeToRefs(notificationsStore)

function closeNotification(id: string) {
  notificationsStore.removeNotification(id)
}
</script>

<template>
  <div class="notifications" v-if="notifications.length">
    <div
      v-for="notification in notifications"
      :key="notification.id"
      :class="['notification', notification.type]"
    >
      <div class="notification-content">
        <span class="notification-message">{{ notification.message }}</span>
        <button class="notification-close" type="button" @click="closeNotification(notification.id)">×</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.notifications {
  position: fixed;
  bottom: 1rem;
  right: 1rem;
  display: grid;
  gap: 0.75rem;
  z-index: 10000;
  max-width: min(360px, calc(100vw - 2rem));
}

.notification {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.95rem 1rem;
  border-radius: 8px;
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.12);
  color: white;
}

.notification.success {
  background: #16a34a;
}

.notification.error {
  background: #dc2626;
}

.notification.info {
  background: #2563eb;
}

.notification-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 0.75rem;
}

.notification-message {
  flex: 1;
  font-size: 0.95rem;
  line-height: 1.4;
}

.notification-close {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  font-size: 1.2rem;
  line-height: 1;
  padding: 0;
}
</style>
