import axios from 'axios';
import { defineStore } from 'pinia';

export type TrustModelInstance = {
  id: string,
  fullTMI: string,
  application: string,
  sessionId: string,
  template: string,
  active: boolean,
  latestVersion: number,
  atls?: {[key: string]: ActualTrustworthinessLevel},
  states?: {[key: string]: TrustModelInstanceState},
  updates?: {[key: string]: TrustModelInstanceUpdate}
};

export type TrustModelInstanceUpdate = any

export type SubjectiveLogicOpinion = {
  belief: number,
  disbelief: number,
  uncertainty: number,
  base_rate: number
};

export type ActualTrustworthinessLevel = {
  TmiID: string,
  Version: number,
  SlResults: {[key: string]: SubjectiveLogicOpinion},
  PpResults: {[key: string]: number},
  TdResults: {[key: string]: number}
};

export type TrustModelInstanceState = {
  Version: number,
  Fingerprint: number,
  Structure: {
    operator: string,
    adjacency_list: {
      sourceNode: string,
      targetNodes: string[]
    }[]
  },
  Values: {
    [key: string]: {
      source: string,
      destination: string,
      opinion: SubjectiveLogicOpinion
    }[]
  },
  RTLs: { [key: string]: SubjectiveLogicOpinion }
};

export const useAppStore = defineStore('app', {
  state: () => ({
    loading: false as (boolean),
    socket: null as (WebSocket|null),
    trustModelInstances: {} as {[key: string]: TrustModelInstance},
  }),

  actions: {
    setSocket(socket: WebSocket|null) {
      this.socket = socket;
    },

    processMessage(msg: any) {
      // this.log.push(transformMessage(data));
      switch (msg.EventType) {
        case 'ATL_REMOVED':
        case 'SESSION_CREATED':
          // ignore
          break;

        case 'TRUST_MODEL_INSTANCE_SPAWNED': {
          const parts = msg.FullTMI.split('/');
          this.trustModelInstances[msg.FullTMI] = {
            id: msg.ID,
            fullTMI: msg.FullTMI,
            application: parts[2],
            sessionId: parts[3],
            template: parts[4],
            active: true,
            latestVersion: msg.Version
          };
          break;
        }

        case 'ATL_UPDATED': {
          const key = msg.FullTMI;
          if (this.trustModelInstances[key]) {
            // ignore update if tmi is not known?
            if (!this.trustModelInstances[key].atls) {
              this.trustModelInstances[key].atls = {};
            }
            this.trustModelInstances[key].atls[msg.Version] = msg.NewATLs;
          } else {
            console.log('[warn] tmi not known', key);
          }
          break;
        }

        case 'TRUST_MODEL_INSTANCE_DELETED': {
          const key = msg.FullTMI;
          if (this.trustModelInstances[key]) {
            this.trustModelInstances[key].active = false;
          }
          break;
        }

        case 'TRUST_MODEL_INSTANCE_UPDATED': {
          const key = msg.FullTMI;
          if (this.trustModelInstances[key]) {
            // ignore update if tmi is not known?
            if (!this.trustModelInstances[key].states) {
              this.trustModelInstances[key].states = {};
            }
            this.trustModelInstances[key].states[msg.Version] = msg;

            if (msg.Update) {
              if (!this.trustModelInstances[key].updates) {
                this.trustModelInstances[key].updates = {};
              }

              if (!Array.isArray(this.trustModelInstances[key].updates[msg.Version])) {
                this.trustModelInstances[key].updates[msg.Version] = [];
              }

              this.trustModelInstances[key].updates[msg.Version].push(msg.Update);
            }
          } else {
            console.log('[warn] tmi not known', key);
          }
          console.log('oi', msg);
          break;
        }

        default:
          console.log('[incoming]', msg.EventType, JSON.stringify(msg, null, 2));
      }
    },

    async fetchTrustModelInstance(application: string, sessionId: string, template: string, id: string, version: string='all') {
      const req = await axios.get(`/api/tmis/${application}/${sessionId}/${template}/${id}/${version}`);
      const key = req.data.fullTMI as string;

      if (!this.trustModelInstances[key]) {
        const parts = req.data.fullTMI.split('/');
        this.trustModelInstances[key] = {
          id: req.data.id,
          fullTMI: req.data.fullTMI,
          application: parts[2],
          sessionId: parts[3],
          template: req.data.template,
          active: req.data.active,
          latestVersion: req.data.latestVersion
        };
      }

      if (!this.trustModelInstances[key].atls) {
        this.trustModelInstances[key].atls = {};
      }

      if (!this.trustModelInstances[key].states) {
        this.trustModelInstances[key].states = {};
      }

      if (!this.trustModelInstances[key].updates) {
        this.trustModelInstances[key].updates = {};
      }

      if (req.data.atls?.Version !== undefined) {
        this.trustModelInstances[key].atls[req.data.atls.Version] = req.data.atls;
      }

      if (req.data.state?.Version !== undefined) {
        this.trustModelInstances[key].states[req.data.state.Version] = req.data.state;

        if (Array.isArray(req.data.updates)) {
          this.trustModelInstances[key].updates[req.data.state.Version] = req.data.updates;
        }
      } else if (typeof(req.data.states) === 'object') {
        this.trustModelInstances[key].states = req.data.states;
        if (typeof(req.data.updates) === 'object') {
          this.trustModelInstances[key].updates = req.data.updates;
        }
        if (typeof(req.data.atls) === 'object') {
          this.trustModelInstances[key].atls = req.data.atls;
        }
      }
    },

    async fetchTrustModelInstances() {
      const req = await axios.get('/api/tmis');
      this.trustModelInstances = Object.fromEntries(
        Object.entries(req.data).map(([key, entry]: [string, any]) => {
          const parts = key.split('/');

          return [entry.fullTMI, {
            id: entry.id,
            fullTMI: entry.fullTMI,
            application: parts[2],
            sessionId: parts[3],
            template: entry.template,
            active: entry.active,
            latestVersion: entry.latestVersion
          }];
        })
      );
    },

    async init() {
      await this.fetchTrustModelInstances();
    }
  }
})
