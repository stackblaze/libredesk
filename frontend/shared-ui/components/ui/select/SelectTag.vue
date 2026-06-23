<template>
  <!-- idk why I named this select tag, should be named multi-select -->
  <TagsInput v-model="tags" class="px-0 gap-0" :displayValue="getLabel">
    <!-- Tags visible to the user -->
    <div class="flex gap-2 flex-wrap items-center px-3">
      <TagsInputItem v-for="tagValue in tags" :key="tagValue" :value="tagValue">
        <TagsInputItemText />
        <TagsInputItemDelete />
      </TagsInputItem>
    </div>

    <!-- Combobox for selecting new tags -->
    <ComboboxRoot
      :model-value="tags"
      v-model:open="open"
      v-model:search-term="searchTerm"
      :filterFunction="filterFunc"
      class="w-full"
    >
      <ComboboxAnchor as-child>
        <ComboboxInput :placeholder="placeholder" as-child>
          <TagsInputInput
            class="w-full px-3"
            :class="tags.length > 0 ? 'mt-2' : ''"
            @keydown.enter.prevent="onEnterKey"
            @blur="handleBlur"
            @click="open = true"
            @input.stop
          />
        </ComboboxInput>
      </ComboboxAnchor>
      <ComboboxPortal>
        <ComboboxContent>
          <CommandList
            position="popper"
            class="w-[--radix-popper-anchor-width] rounded-md mt-2 border bg-popover text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2"
          >
            <CommandEmpty v-if="!canCreateFromSearch">
              {{ emptyHint }}
            </CommandEmpty>
            <CommandGroup>
              <CommandItem
                v-if="canCreateFromSearch"
                :value="`__create__${searchTerm.trim()}`"
                @select="createFromSearch"
              >
                {{ t('tag.createNamed', { name: searchTerm.trim() }) }}
              </CommandItem>
              <CommandItem
                v-for="item in visibleOptions"
                :key="item.value"
                :value="item.value"
                @select="handleSelect"
              >
                {{ item.label }}
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </ComboboxContent>
      </ComboboxPortal>
    </ComboboxRoot>
  </TagsInput>
</template>

<script setup>
import { CommandEmpty, CommandGroup, CommandItem, CommandList } from '../command'
import {
  TagsInput,
  TagsInputInput,
  TagsInputItem,
  TagsInputItemDelete,
  TagsInputItemText
} from '../tags-input'
import {
  ComboboxAnchor,
  ComboboxContent,
  ComboboxInput,
  ComboboxPortal,
  ComboboxRoot
} from 'radix-vue'
import { computed, ref } from 'vue'
import { useField } from 'vee-validate'
import { useI18n } from 'vue-i18n'

const RENDER_CAP = 200

const tags = defineModel({
  required: false,
  default: () => []
})

const props = defineProps({
  name: {
    type: String,
    required: false,
    default: 'tags'
  },
  placeholder: {
    type: String,
    default: 'Select...'
  },
  items: {
    type: Array,
    required: true,
    validator: (value) => value.every((item) => 'label' in item && 'value' in item)
  },
  allowCreate: {
    type: Boolean,
    default: true
  }
})

const { t } = useI18n()

const { handleBlur } = useField(() => props.name, undefined, {
  initialValue: tags.value
})

const open = ref(false)
const searchTerm = ref('')

const filteredOptions = computed(() => {
  const available = props.items.filter((item) => !tags.value.includes(item.value))

  if (!searchTerm.value) return available

  return available.filter((item) =>
    item.label.toLowerCase().includes(searchTerm.value.toLowerCase())
  )
})

const visibleOptions = computed(() => filteredOptions.value.slice(0, RENDER_CAP))

const canCreateFromSearch = computed(() => {
  if (!props.allowCreate) return false
  const term = searchTerm.value.trim()
  if (!term) return false
  const lower = term.toLowerCase()
  if (tags.value.some((tag) => tag.toLowerCase() === lower)) return false
  return !props.items.some((item) => item.label.toLowerCase() === lower)
})

const emptyHint = computed(() => {
  if (props.allowCreate) return t('tag.typeToCreate')
  return t('globals.messages.noResultsFound')
})

const getLabel = (value) => {
  const item = props.items.find((item) => item.value === value)
  return item?.label || value
}

const addTagValue = (value) => {
  if (!value || tags.value.includes(value)) return
  tags.value = [...tags.value, value]
  searchTerm.value = ''
  if (filteredOptions.value.length === 0) {
    open.value = false
  }
}

const createFromSearch = () => {
  addTagValue(searchTerm.value.trim())
}

const onEnterKey = () => {
  if (canCreateFromSearch.value) {
    createFromSearch()
    return
  }
  if (visibleOptions.value.length === 1) {
    addTagValue(visibleOptions.value[0].value)
  }
}

const handleSelect = (event) => {
  const selectedValue = event.detail.value
  if (selectedValue) {
    addTagValue(selectedValue)
  }
}

const filterFunc = (remainingItemValues, term) => {
  const remainingItems = props.items.filter((item) => remainingItemValues.includes(item.value))
  return remainingItems
    .filter((item) => item.label.toLowerCase().includes(term.toLowerCase()))
    .map((item) => item.value)
}
</script>
