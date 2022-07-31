import {Readable, Writable, writable} from 'svelte/store';
import rpc from './rpc'

type StartStopNotifier<K, V> = (opt: {
    key: K,
    resolve: (v: V) => void,
    reject: (e: Error) => void,
}) => (() => void) | void;
type Key = string | number | Readonly<any>

interface Entry<T extends Keyed> {
    working: boolean
    value?: T
    error?: Error
}

interface Keyed {
    key(): Key
}

class Actionable<T> {
    // this method should get setup in the createStore method
    action = <P, R>(method: string, param: P): Promise<R> => {
        return Promise.reject(new Error("instance has not been initialized properly"))
    }
}

const createStore = <K extends Key, V extends Keyed>(
    clazz: new () => V,
    notificationMethod: string,
    startStop?: StartStopNotifier<K, V>
) => {
    const db: { [key: string]: Writable<Entry<V>> } = {};

    const set = (v: any, error?: Error) => {
        const inst = Object.assign(new clazz(), v)

        if (inst instanceof Actionable) {
            inst.action = <P, R>(method: string, param: P): Promise<R> => {
                return rpc.send<P, R>(method, param)
            }
        }

        const key = inst.key()
        if (key) get(key).set({working: false, value: inst, error});
    }

    const get = (key: K): Writable<Entry<V>> => {
        const k = JSON.stringify(key);
        if (db[k] === undefined) {
            db[k] = writable<Entry<V>>({working: false}, () => {
                if (startStop) {
                    const resolve = (value: V) => {
                        set(value)
                        // db[k].set({working: false, value, error: undefined})
                    }
                    const reject = (error: Error) => {
                        set(undefined, error)
                        // db[k].set({working: false, value: undefined, error})
                    }
                    db[k].set({working: true, value: undefined, error: undefined})
                    const stop = startStop({key, resolve, reject})
                    if (stop) {
                        return stop
                    }
                }
            });
        }
        return db[k]
    }

    rpc.handleNotification<V>(notificationMethod, set)

    const bind = (key: K): Readable<Entry<V>> => {
        const store = get(key)
        return {
            subscribe: store.subscribe,
        }
    }

    return {
        bind,
    };
}

export type{Key, Keyed}
export {Actionable}
export default createStore