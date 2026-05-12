import { ref } from 'vue'
import { defineStore } from 'pinia'

export type NotificationType = 'success' | 'error' | 'info'

export type Notification = {
  id: string
  message: string
  type: NotificationType
}

export const useNotificationsStore = defineStore('notifications', () => {
  const notifications = ref<Notification[]>([])

  function removeNotification(id: string) {
    notifications.value = notifications.value.filter((notification) => notification.id !== id)
  }

  function updateNotification(id: string, message: string) {
    notifications.value = notifications.value.map((notification) => {
      if (notification.id === id) {
        return { ...notification, message }
      }
      return notification
    })
  }

  function notify(message: string, type: NotificationType = 'success', timeout = 4000) {
    const id = `${Date.now()}-${Math.random().toString(16).slice(2)}`
    notifications.value.push({ id, message, type })

    window.setTimeout(() => removeNotification(id), timeout)
    return id
  }

  function notifySuccess(message: string, timeout = 4000) {
    return notify(message, 'success', timeout)
  }

  function notifyError(message: string, timeout = 6000) {
    return notify(message, 'error', timeout)
  }

  function notifyInfo(message: string, timeout = 4000) {
    return notify(message, 'info', timeout)
  }

  return {
    notifications,
    notify,
    notifySuccess,
    notifyError,
    notifyInfo,
    removeNotification,
    updateNotification,
  }
})
