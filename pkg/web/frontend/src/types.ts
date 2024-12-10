import { InjectionKey, Ref } from 'vue';

import AlertDialog from '@/components/AlertDialog.vue';

export const Dialog: InjectionKey<Ref<InstanceType<typeof AlertDialog> | null>> = Symbol('AlertDialog');

export type AlertType = 'alert' | 'prompt' | 'select' | 'confirm';

export type AlertOptions = {
  color?: string,
  width?: number,
  type?: AlertType,
  wrap?: boolean,
  closable?: boolean,
  icon?: string|null,
  prependIcon?: string|null,
  hint?: string,
  label?: string,
  items?: any[],
  itemText?: any,
  itemValue?: any,
  multiple?: boolean,
  description?: string,
  confirmText?: string,
  confirmIcon?: string|null,
  confirmValue?: any,
  cancelText?: string,
  cancelIcon?: string|null,
  cancelValue?: any,
}

export type FilterSettings = {
  type?: AlertType,
  items?: any[],
  itemText?: any,
  itemValue?: any,
};

export type Column = {
  filterable?: boolean,
  filterSettings?: FilterSettings,
  maxWidth?: string,
  title: string,
  key: string,
}

export type SortItem = {
  key: string,
  order?: boolean | 'asc' | 'desc'
}
