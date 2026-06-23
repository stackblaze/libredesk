<template>
  <div :class="layout === 'sidebar' ? 'space-y-0' : 'space-y-2'">
    <FormField name="contact_email">
      <FormItem :class="fieldClass">
        <FormLabel :class="labelClass">{{ $t('globals.terms.email') }}</FormLabel>
        <FormControl>
          <Input
            data-new-ticket-email
            type="email"
            :placeholder="t('conversation.searchContact')"
            v-model="emailQuery"
            @input="handleSearchContacts"
            @keydown="handleSearchKeydown"
            autocomplete="off"
          />
        </FormControl>
        <FormMessage />

        <div
          v-if="searchResults.length"
          class="absolute left-0 right-0 z-50 mt-1 rounded-md border bg-popover p-1 text-popover-foreground shadow-md"
        >
          <ul class="max-h-60 overflow-y-auto" role="listbox">
            <li
              v-for="(contact, index) in searchResults"
              :key="contact.email"
              @click="selectContact(contact)"
              role="option"
              :aria-selected="index === highlightedIndex"
              class="relative flex cursor-pointer select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors duration-200"
              :class="
                index === highlightedIndex
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent hover:text-accent-foreground'
              "
            >
              <div>
                <p class="font-medium">{{ contact.first_name }} {{ contact.last_name }}</p>
                <p class="text-xs text-muted-foreground">{{ contact.email }}</p>
              </div>
            </li>
          </ul>
        </div>
      </FormItem>
    </FormField>

    <template v-if="layout === 'sidebar'">
      <FormField v-slot="{ componentField }" name="first_name">
        <FormItem :class="fieldClass">
          <FormLabel :class="labelClass">{{ $t('globals.terms.firstName') }}</FormLabel>
          <FormControl>
            <Input type="text" v-bind="componentField" :disabled="!!selectedContact" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="last_name">
        <FormItem :class="fieldClass">
          <FormLabel :class="labelClass">{{ $t('globals.terms.lastName') }}</FormLabel>
          <FormControl>
            <Input type="text" v-bind="componentField" :disabled="!!selectedContact" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="subject">
        <FormItem :class="fieldClass">
          <FormLabel :class="labelClass">{{ $t('globals.terms.subject') }}</FormLabel>
          <FormControl>
            <Input type="text" v-bind="componentField" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="inbox_id">
        <FormItem :class="fieldClass">
          <FormLabel :class="labelClass">{{ $t('globals.terms.inbox') }}</FormLabel>
          <FormControl>
            <Select v-bind="componentField">
              <SelectTrigger>
                <SelectValue :placeholder="t('placeholders.selectInbox')" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem
                    v-for="option in inboxStore.emailOptions"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="team_id">
        <FormItem :class="fieldClass">
          <FormLabel :class="labelClass">
            {{ $t('actions.assignTeam') }} ({{ $t('globals.terms.optional') }})
          </FormLabel>
          <FormControl>
            <SelectComboBox
              v-bind="componentField"
              :items="[{ value: 'none', label: t('globals.terms.none') }, ...teamStore.options]"
              :placeholder="t('placeholders.selectTeam')"
              type="team"
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="agent_id">
        <FormItem :class="fieldClass">
          <FormLabel :class="labelClass">
            {{ $t('actions.assignAgent') }} ({{ $t('globals.terms.optional') }})
          </FormLabel>
          <FormControl>
            <SelectComboBox
              v-bind="componentField"
              :items="[{ value: 'none', label: t('globals.terms.none') }, ...uStore.options]"
              :placeholder="t('placeholders.selectAgent')"
              type="user"
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

    </template>

    <template v-else>
      <div class="grid grid-cols-2 gap-4">
        <FormField v-slot="{ componentField }" name="first_name">
          <FormItem>
            <FormLabel>{{ $t('globals.terms.firstName') }}</FormLabel>
            <FormControl>
              <Input type="text" v-bind="componentField" :disabled="!!selectedContact" required />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="last_name">
          <FormItem>
            <FormLabel>{{ $t('globals.terms.lastName') }}</FormLabel>
            <FormControl>
              <Input type="text" v-bind="componentField" :disabled="!!selectedContact" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <FormField v-slot="{ componentField }" name="subject">
          <FormItem>
            <FormLabel>{{ $t('globals.terms.subject') }}</FormLabel>
            <FormControl>
              <Input type="text" v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="inbox_id">
          <FormItem>
            <FormLabel>{{ $t('globals.terms.inbox') }}</FormLabel>
            <FormControl>
              <Select v-bind="componentField">
                <SelectTrigger>
                  <SelectValue :placeholder="t('placeholders.selectInbox')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="option in inboxStore.emailOptions"
                      :key="option.value"
                      :value="option.value"
                    >
                      {{ option.label }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <FormField v-slot="{ componentField }" name="team_id">
          <FormItem>
            <FormLabel>
              {{ $t('actions.assignTeam') }} ({{ $t('globals.terms.optional') }})
            </FormLabel>
            <FormControl>
              <SelectComboBox
                v-bind="componentField"
                :items="[{ value: 'none', label: t('globals.terms.none') }, ...teamStore.options]"
                :placeholder="t('placeholders.selectTeam')"
                type="team"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="agent_id">
          <FormItem>
            <FormLabel>
              {{ $t('actions.assignAgent') }} ({{ $t('globals.terms.optional') }})
            </FormLabel>
            <FormControl>
              <SelectComboBox
                v-bind="componentField"
                :items="[{ value: 'none', label: t('globals.terms.none') }, ...uStore.options]"
                :placeholder="t('placeholders.selectAgent')"
                type="user"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Input } from '@shared-ui/components/ui/input'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@shared-ui/components/ui/form'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@shared-ui/components/ui/select'
import SelectComboBox from '@/components/combobox/SelectCombobox.vue'

const emailQuery = defineModel('emailQuery', { type: String, default: '' })

defineProps({
  layout: {
    type: String,
    default: 'grid'
  },
  inboxStore: { type: Object, required: true },
  uStore: { type: Object, required: true },
  teamStore: { type: Object, required: true },
  searchResults: { type: Array, required: true },
  highlightedIndex: { type: Number, required: true },
  selectedContact: { type: Object, default: null },
  handleSearchContacts: { type: Function, required: true },
  handleSearchKeydown: { type: Function, required: true },
  selectContact: { type: Function, required: true }
})

const { t } = useI18n()

const fieldClass = computed(() => 'prop-field relative')
const labelClass = computed(() => 'prop-label')
</script>
