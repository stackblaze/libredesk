import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { handleHTTPError } from '@shared-ui/utils/http.js'
import { useEmitter } from '../composables/useEmitter'
import { EMITTER_EVENTS } from '../constants/emitterEvents'
import api from '../api'

export const useTagStore = defineStore('tags', () => {
    const tags = ref([])
    const emitter = useEmitter()
    const tagNames = computed(() => tags.value.map(tag => tag.name))
    const tagOptions = computed(() => tags.value.map(tag => ({
        label: tag.name,
        value: String(tag.id),
    })))

    const fetchTags = async ({ force = false } = {}) => {
        if (!force && tags.value.length > 0) return
        try {
            const response = await api.getTags()
            tags.value = response?.data?.data || []
        } catch (error) {
            emitter.emit(EMITTER_EVENTS.SHOW_TOAST, {
                variant: 'destructive',
                description: handleHTTPError(error).message
            })
        }
    }

    const invalidateTags = () => {
        tags.value = []
    }

    return {
        tags,
        tagOptions,
        tagNames,
        fetchTags,
        invalidateTags,
    }
})