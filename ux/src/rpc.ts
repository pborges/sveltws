import {nanoid} from "nanoid";

interface Message<T> {
    method: string
    id?: string
    params?: T
    result?: T
    error?: string
}

class Client {
    private ws: WebSocket;
    private openRequests: { [key: string]: (result: Message<any>) => void } = {};
    private notificationHandlers: { [key: string]: (arg: any) => void } = {};

    constructor(addr: string) {
        console.log("create rpc")
        this.ws = new WebSocket(addr);

        this.ws.onmessage = this.onmessage
        this.ws.onopen = () => console.log("open")
        this.ws.onclose = () => console.log("close")
        this.ws.onerror = console.log
    }

    private onmessage = (evt: MessageEvent) => {
        const data: Message<any> = JSON.parse(evt.data);
        console.log("[WS] <<", data);
        if (data.id !== undefined) {
            const h = this.openRequests[data.id];
            if (h) {
                h(data);
                this.openRequests[data.id] = undefined
            }
        } else {
            const h = this.notificationHandlers[data.method];
            if (h) h(data.params);
        }
    }

    public handleNotification = <T>(method: string, cb: (arg: T) => void) => {
        this.notificationHandlers[method] = cb;
    }

    public send = <P, R>(method: string, params: P): Promise<R> => {
        return new Promise((resolve, reject) => {
            const req: Message<any> = {
                id: nanoid(),
                method,
                params,
            }
            this.openRequests[req.id] = (res: Message<R>) => {
                if (res.error === undefined) {
                    resolve(res.result)
                } else {
                    reject(new Error(res.error))
                }
            }
            console.log("[WS] >>", req);
            this.ws.send(JSON.stringify(req));
        })
    }
}

const rpc = new Client("ws://localhost:8081/ws")

export default rpc;