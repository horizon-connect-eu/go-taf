<template>
    <v-dialog v-model="$dialog" :max-width="$options.width" @keydown.esc="cancel" scrollable persistent>
        <v-card>
            <v-toolbar dark :color="$options.color" dense text max-height="64px">
                <v-toolbar-title>
                    <v-icon v-if="$options.icon" start :icon="$options.icon" />
                    {{ $title }}
                </v-toolbar-title>

                <v-spacer v-if="$options.closable" />

                <v-btn icon @click="close()" v-if="$options.closable">
                    <v-icon icon="mdi-close" />
                </v-btn>
            </v-toolbar>

            <v-card-text v-if="$options.type !== 'prompt' && $options.type !== 'select'" v-show="!!$message" :style="$options.wrap ? 'white-space: pre-wrap;' : ''" class="pt-5">
              {{ $message }}
            </v-card-text>

            <v-card-text v-if="$options.type === 'prompt'" class="pt-5">
                <v-text-field ref="input" v-model="$message" :label="$options.label" :hint="$options.hint" :persistent-hint="!!$options.hint" :prepend-icon="$options.prependIcon || undefined" autofocus outlined @keyup.enter="agree" />
            </v-card-text>

            <v-card-text v-if="$options.type === 'select'" class="pt-5">
                <p :style="$options.wrap ? 'white-space: pre-wrap;' : ''">{{ $options.description }}</p>
                <v-select ref="dropdown" v-model="$message" :item-title="$options.itemText" :item-value="$options.itemValue" :items="$options.items" :label="$options.label" :hint="$options.hint" :persistent-hint="!!$options.hint" :prepend-icon="$options.prependIcon || undefined" outlined autofocus @keyup.enter="agree">
                  <template v-slot:item="{ props, item }">
                    <v-list-item v-bind="props" :base-color="item.raw.color" :subtitle="item.raw.subtitle"></v-list-item>
                  </template>
                </v-select>
            </v-card-text>

            <v-card-actions class="pt-0">
                <v-col cols="12" class="text-right">
                    <v-btn v-if="$options.type !== 'alert' && $options.cancelText" @click="cancel" :color="$options.cancelColor">
                        <v-icon start v-if="$options.cancelIcon" :icon="$options.cancelIcon" />
                        {{ $options.cancelText }}
                    </v-btn>
                    <v-btn v-if="$options.type !== 'alert' && $options.confirmText" @click="agree" :color="$options.confirmColor">
                        <v-icon start v-if="$options.confirmIcon" :icon="$options.confirmIcon" />
                        {{ $options.confirmText }}
                    </v-btn>
                    <v-btn v-if="$options.type === 'alert'" text="text" @click="cancel">Ok</v-btn>
                </v-col>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { VSelect, VTextField } from 'vuetify/components';

import { AlertOptions } from '@/types';

const defaults: AlertOptions = {
  color: 'primary',
  width: 520,
  wrap: true,
  closable: false,
  icon: null,
  prependIcon: null,
  hint: '',
  label: '',
  items: [],
  multiple: false,
  itemText: null,
  itemValue: null,
  description: '',
  cancelText: 'Cancel',
  confirmText: 'Yes',
  confirmValue: true,
  cancelValue: false,
};

const input = ref<null|VTextField>(null);
const dropdown = ref<null|VSelect>(null);

const $dialog = ref<boolean>(false);
const $title = ref<string>('');
const $message = ref<string>('');
const $options = ref({...defaults});
const $promise = ref<null|Promise<any>>(null);
const $resolve = ref<null|((value?: any | PromiseLike<any>) => void)>(null);

async function open(title: string, message: string, options: AlertOptions): Promise<any> {
  $options.value = Object.assign({...defaults}, options);

  $title.value = title;
  $message.value = message;

  $dialog.value = true;

  if ($promise.value) {
    await $promise.value;
  }

  $promise.value = new Promise((resolve) => {
    $resolve.value = resolve;
  });

  setTimeout(() => {
    console.log(dropdown.value)
    switch ($options.value.type) {
      case 'prompt':
        input?.value?.focus();
        break;

      case 'select':
        dropdown?.value?.focus();
        break;
    }
  }, 50);

  return $promise.value;
}

function alert(title: string, message: string = '', options: AlertOptions = {}): Promise<any> {
  return open(title, message, Object.assign({ type: 'alert' }, options));
}

function prompt(title: string, input: string = '', options: AlertOptions = {}): Promise<any> {
  return open(title, input, Object.assign({ type: 'prompt' }, options));
}

function confirm(title: string, input: string = '', options: AlertOptions = {}): Promise<any> {
  return open(title, input, Object.assign({ type: 'confirm' }, options));
}

function select(title: string, input: string = '', options: AlertOptions = {}): Promise<any> {
  return open(title, input, Object.assign({ type: 'select' }, options));
}

function close(value: any = null) {
  if ($resolve.value) {
    $resolve.value(value);
    $promise.value = null;
    $dialog.value = false;
  }
}

function agree() {
  switch ($options.value.type) {
    case 'prompt':
      return close($message.value);

    case 'select':
      return close($message.value);

    default:
      return close($options.value?.confirmValue ?? true);
  }
}

function cancel() {
  close($options.value?.cancelValue ?? false);
}

defineExpose({
  open,
  alert,
  prompt,
  confirm,
  select
});
</script>
