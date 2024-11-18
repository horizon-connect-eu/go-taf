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
  states: {[key: string]: TrustModelInstanceState},
  atls: {[key: string]: ActualTrustworthinessLevel},
  updates: {[key: string]: TrustModelInstanceUpdate}
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

export type Session = {
  Client:   string,
  IsActive: boolean,
  TMIs:     string[],
  Template: string,
};

export const useAppStore = defineStore('app', {
  state: () => ({
    loading: false as (boolean),
    socket: null as (WebSocket|null),
    sessions: {} as {[key: string]: Session},
    trustModelInstances: {} as {[key: string]: TrustModelInstance},
  }),

  actions: {
    setSocket(socket: WebSocket|null) {
      this.socket = socket;
    },

    processMessage(msg: any) {
      // process events by the taf and update state by replicating the go logic
      console.log('[ws:recv]', msg);

      switch (msg.EventType) {
        case 'SESSION_CREATED':
          /*
            func (s *State) handleSessionCreatedEvent(event listener.SessionCreatedEvent) {
              s.sessions[event.SessionID] = &sessionState{
                Client:   event.ClientID,
                IsActive: true,
                TMIs:     make([]string, 0),
                Template: event.TrustModelTemplate,
              }
            }
          */
          if (msg.SessionID) {
            this.sessions[msg.SessionID] = {
              Template: msg.TrustModelTemplate,
              Client: msg.ClientID,
              IsActive: true,
              TMIs: [],
            };
          }
          break;

        case 'SESSION_TORNDOWN':
          /*
            func (s *State) handleSessionTorndownEvent(event listener.SessionTorndownEvent) {
              if _, exists := s.sessions[event.SessionID]; exists {
                s.sessions[event.SessionID].IsActive = false
              }
            }
          */
          if (msg.SessionID && this.sessions[msg.SessionID]) {
            this.sessions[msg.SessionID].IsActive = false;
          }
          break;

        case 'ATL_REMOVED':
          // ignore
          break;

        case 'TRUST_MODEL_INSTANCE_SPAWNED': {
          /*
            func (s *State) handleTMISpawned(event listener.TrustModelInstanceSpawnedEvent) {
              fullTMI := event.FullTMI

              _, sessionID, _, _ := core.SplitFullTMIIdentifier(fullTMI)
              if _, exists := s.sessions[sessionID]; exists {
                s.sessions[sessionID].TMIs = append(s.sessions[sessionID].TMIs, fullTMI)
              }

              s.tmis[fullTMI] = &tmiMetaState{
                IsActive:      true,
                LatestVersion: 0,
                Update:        make(map[int][]core.Update),
                States:        make(map[int]tmiState),
                Template:      event.Template,
                ID:            event.ID,
                FullTMI:       event.FullTMI,
                ATLs:          make(map[int]core.AtlResultSet),
              }
              s.tmis[fullTMI].States[event.Version] = tmiState{
                Version:     event.Version,
                Fingerprint: event.Fingerprint,
                Structure:   event.Structure,
                Values:      event.Values,
                RTLs:        event.RTLs,
              }

            }
          */
          const parts = msg.FullTMI.split('/');

          if (this.sessions[parts[3]]) {
            this.sessions[parts[3]].TMIs.push(msg.FullTMI);
          }

          this.trustModelInstances[msg.FullTMI] = {
            id: msg.ID,
            fullTMI: msg.FullTMI,
            application: parts[2],
            sessionId: parts[3],
            template: parts[4],
            active: true,
            latestVersion: msg.Version,
            updates: {},
            states: {
              [String(msg.Version)]: {
                Version: msg.Version,
                Fingerprint: msg.Fingerprint,
                Structure: msg.Structure,
                Values: msg.Values,
                RTLs: msg.RTLs
              }
            },
            atls: {}
          };
          break;
        }

        case 'ATL_UPDATED': {
          /*
            func (s *State) handleATLUpdatedEvent(event listener.ATLUpdatedEvent) {
              fullTMI := event.FullTMI
              _, exists := s.tmis[fullTMI]
              if !exists {
                return
              } else {
                s.logger.Warn(fmt.Sprintf("%+v", event.NewATLs))
              }
              s.tmis[fullTMI].ATLs[event.NewATLs.Version()] = event.NewATLs
            }
          */
          const key = msg.FullTMI;
          if (this.trustModelInstances[key]) {
            // ignore update if tmi is not known?
            if (!this.trustModelInstances[key].atls) {
              this.trustModelInstances[key].atls = {};
            }
            this.trustModelInstances[key].atls[msg.NewATLs.Version] = msg.NewATLs;
          } else {
            console.log('[warn] tmi not known', key);
          }
          break;
        }

        case 'TRUST_MODEL_INSTANCE_DELETED': {
          /*
            func (s *State) handleTMIDeleted(event listener.TrustModelInstanceDeletedEvent) {
              fullTMI := event.FullTMI
              s.tmis[fullTMI].IsActive = false
            }
          */

          const key = msg.FullTMI;
          if (this.trustModelInstances[key]) {
            this.trustModelInstances[key].active = false;
          }
          break;
        }

        case 'TRUST_MODEL_INSTANCE_UPDATED': {
          /*
            func (s *State) handleTMIUpdated(event listener.TrustModelInstanceUpdatedEvent) {
              fullTMI := event.FullTMI
              // If there exists already an entry for that version, this means we have received another update that yields the same
              // version number. This means that the second update has failed to increase the version number and can be ignored.
              _, exists := s.tmis[fullTMI].States[event.Version]
              if exists {
                s.tmis[fullTMI].Update[event.Version+1] = []core.Update{event.Update}
                return
              }
              s.tmis[fullTMI].States[event.Version] = tmiState{
                Version:     event.Version,
                Fingerprint: event.Fingerprint,
                Structure:   event.Structure,
                Values:      event.Values,
                RTLs:        event.RTLs,
              }
              if s.tmis[fullTMI].Update[event.Version] == nil {
                s.tmis[fullTMI].Update[event.Version] = make([]core.Update, 0)
              }
              s.tmis[fullTMI].Update[event.Version] = append(s.tmis[fullTMI].Update[event.Version], event.Update)
              s.tmis[fullTMI].LatestVersion = event.Version
            }
          */

          const key = msg.FullTMI;
          if (this.trustModelInstances[key]) {
            // ignore update if tmi is not known?
            if (!this.trustModelInstances[key].states) {
              this.trustModelInstances[key].states = {};
            }

            if (msg.Update) {
              if (!this.trustModelInstances[key].updates) {
                this.trustModelInstances[key].updates = {};
              }

              if (this.trustModelInstances[key].states[msg.Version]) {
                if (!Array.isArray(this.trustModelInstances[key].updates[msg.Version + 1])) {
                  this.trustModelInstances[key].updates[msg.Version + 1] = [];
                }

                this.trustModelInstances[key].updates[msg.Version + 1].push(msg.Update);
                return;
              }

              if (!Array.isArray(this.trustModelInstances[key].updates[msg.Version])) {
                this.trustModelInstances[key].updates[msg.Version] = [];
              }

              this.trustModelInstances[key].updates[msg.Version].push(msg.Update);
            }

            this.trustModelInstances[key].states[msg.Version] = msg;
          } else {
            console.log('[warn] tmi not known', key);
          }
          break;
        }

        default:
          console.log('[incoming]', msg.EventType, JSON.stringify(msg, null, 2));
      }
    },

    async fetchTrustModelInstance(application: string, sessionId: string, template: string, id: string, version: string='all') {
      const res = await axios.get(`/api/tmis/${application}/${sessionId}/${template}/${id}/${version}`);
      const key = res.data.fullTMI as string;

      if (!this.trustModelInstances[key]) {
        const parts = res.data.fullTMI.split('/');
        console.log(res.data);
        this.trustModelInstances[key] = {
          id: res.data.id,
          fullTMI: res.data.fullTMI,
          application: parts[2],
          sessionId: parts[3],
          template: res.data.template,
          active: res.data.active,
          latestVersion: res.data.latestVersion,
          atls: {},
          states: {},
          updates: {},
        };
      }

      if (res.data.atls?.Version !== undefined) {
        this.trustModelInstances[key].atls[res.data.atls.Version] = res.data.atls;
      }

      if (res.data.state?.Version !== undefined) {
        this.trustModelInstances[key].states[res.data.state.Version] = res.data.state;

        if (Array.isArray(res.data.updates)) {
          this.trustModelInstances[key].updates[res.data.state.Version] = res.data.updates;
        }
      } else if (typeof(res.data.states) === 'object') {
        this.trustModelInstances[key].states = res.data.states;
        if (typeof(res.data.updates) === 'object') {
          this.trustModelInstances[key].updates = res.data.updates;
        }
        if (typeof(res.data.atls) === 'object') {
          this.trustModelInstances[key].atls = res.data.atls;
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
            latestVersion: entry.latestVersion,
            states: {},
            updates: {},
            atls: {}
          }];
        })
      );
    },

    async init() {
      await this.fetchTrustModelInstances();
    }
  }
})
