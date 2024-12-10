import { InjectionKey } from 'vue';

import { AlertOptions } from './types';

export const Dialog: InjectionKey<{
  alert: (title: string, message: string, options: AlertOptions) => Promise<any>,
  prompt: (title: string, message: string, options: AlertOptions) => Promise<any>,
  select: (title: string, message: string, options: AlertOptions) => Promise<any>,
  confirm: (title: string, message: string, options: AlertOptions) => Promise<any>
}> = Symbol('AlertDialog');
