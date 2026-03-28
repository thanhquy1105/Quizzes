

export enum MsgType {
    Unknown = 0,
    Connect = 1,
    Connack = 2,
    Request = 3,
    Resp = 4,
    Heartbeat = 5,
    Message = 6,
    BatchMessage = 7,
}

const MAGIC = "BTASKEE";
const MAGIC_BYTES = new TextEncoder().encode(MAGIC);
const MAGIC_LEN = MAGIC_BYTES.length;
const TYPE_LEN = 1;
const CONTENT_LEN = 4;
const HEADER_MIN_LEN = MAGIC_LEN + TYPE_LEN;
const HEADER_FULL_LEN = MAGIC_LEN + TYPE_LEN + CONTENT_LEN;

export type MessageHandler = (msgType: MsgType, data: Uint8Array) => void;

export class WsClient {
    private ws: WebSocket | null = null;
    private handlers: Map<string, (data: any) => void> = new Map();
    private onMessageCallback: MessageHandler | null = null;
    private heartbeatInterval: any = null;

    constructor(private url: string) { }

    private connectResolver: (() => void) | null = null;
    private connectRejecter: ((err: any) => void) | null = null;

    connect(username: string, token: string): Promise<void> {
        return new Promise((resolve, reject) => {
            this.connectResolver = resolve;
            this.connectRejecter = reject;
            this.ws = new WebSocket(this.url);
            this.ws.binaryType = "arraybuffer";

            this.ws.onopen = () => {
                console.log("Connected to WsServer, authenticating...");
                this.startHeartbeat();
                this.sendConnect(username, token);
            };

            this.ws.onerror = (err) => {
                console.error("WebSocket error:", err);
                if (this.connectRejecter) this.connectRejecter(err);
                reject(err);
            };

            this.ws.onmessage = (event) => {
                console.log("websocket on data")
                this.handleData(event.data);
            };

            this.ws.onclose = () => {
                console.log("Disconnected from WsServer");
                this.stopHeartbeat();
            };
        });
    }

    private sendConnect(username: string, token: string) {
        const requestId = BigInt(Math.floor(Math.random() * 1000000));
        const usernameBytes = new TextEncoder().encode(username);
        const tokenBytes = new TextEncoder().encode(token);
        const bodyBytes = new Uint8Array(0); // Empty body for now

        // Connect binary: Id(8) + UsernameLen(2) + Username + TokenLen(2) + Token + BodyLen(4) + Body (LittleEndian)
        const totalLen = 8 + 2 + usernameBytes.length + 2 + tokenBytes.length + 4 + bodyBytes.length;
        const buffer = new ArrayBuffer(totalLen);
        const view = new DataView(buffer);
        const uint8 = new Uint8Array(buffer);

        let offset = 0;
        view.setBigUint64(offset, requestId, true);
        offset += 8;

        view.setUint16(offset, usernameBytes.length, true);
        offset += 2;
        uint8.set(usernameBytes, offset);
        offset += usernameBytes.length;

        view.setUint16(offset, tokenBytes.length, true);
        offset += 2;
        uint8.set(tokenBytes, offset);
        offset += tokenBytes.length;

        view.setUint32(offset, bodyBytes.length, true);
        offset += 4;
        uint8.set(bodyBytes, offset);

        this.send(MsgType.Connect, uint8);
    }

    private handleData(buffer: ArrayBuffer) {
        const uint8 = new Uint8Array(buffer);
        const view = new DataView(buffer);

        if (uint8.length < HEADER_MIN_LEN) return;

        // Check MAGIC
        for (let i = 0; i < MAGIC_LEN; i++) {
            if (uint8[i] !== MAGIC_BYTES[i]) return;
        }

        const msgType = view.getUint8(MAGIC_LEN) as MsgType;

        if (msgType === MsgType.Heartbeat) {
            if (this.onMessageCallback) this.onMessageCallback(msgType, new Uint8Array(0));
            return;
        }

        if (uint8.length < HEADER_FULL_LEN) return;
        const dataLen = view.getUint32(MAGIC_LEN + TYPE_LEN, false); // Big Endian in protocol envelope
        const data = uint8.slice(HEADER_FULL_LEN, HEADER_FULL_LEN + dataLen);

        if (msgType === MsgType.Connack) {
            this.handleConnack(data);
            return;
        }

        if (msgType === MsgType.Resp) {
            this.handleResponse(data);
        }

        if (msgType === MsgType.Message) {
            this.handleMessage(data);
            return;
        }

        if (this.onMessageCallback) {
            this.onMessageCallback(msgType, data);
        }
    }

    private handleMessage(data: Uint8Array) {
        if (data.length < 24) return; // Id(8) + MsgType(4) + Timestamp(8) + ContentLen(4)
        const view = new DataView(data.buffer, data.byteOffset, data.byteLength);

        // Skip Id(8), MsgType(4), Timestamp(8)
        const contentLen = view.getUint32(20, true);
        const content = data.slice(24, 24 + contentLen);

        if (this.onMessageCallback) {
            this.onMessageCallback(MsgType.Message, content);
        }
    }

    private handleConnack(data: Uint8Array) {
        if (data.length < 9) return; // Id(8) + Status(1)
        const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
        const status = view.getUint8(8);

        if (status === 0) { // StatusOK
            console.log("Authenticated successfully!");
            if (this.connectResolver) this.connectResolver();
        } else {
            console.error("Authentication failed with status:", status);
            if (this.connectRejecter) this.connectRejecter(new Error(`Authentication failed: ${status}`));
        }
    }

    private handleResponse(data: Uint8Array) {
        if (data.length < 21) return; // Id(8) + Status(1) + Timestamp(8) + BodyLen(4)
        const view = new DataView(data.buffer, data.byteOffset, data.byteLength);

        let offset = 0;
        const id = view.getUint32(offset, true) + (view.getUint32(offset + 4, true) * 0x100000000); // Read uint64 as two uint32
        offset += 8;

        const status = view.getUint8(offset);
        offset += 1;

        const timestamp = view.getUint32(offset, true) + (view.getUint32(offset + 4, true) * 0x100000000);
        offset += 8;

        const bodyLen = view.getUint32(offset, true);
        offset += 4;

        const bodyBytes = data.slice(offset, offset + bodyLen);
        const body = new TextDecoder().decode(bodyBytes);

        let parsedBody = body;
        try {
            parsedBody = JSON.parse(body);
        } catch (e) { }
        console.log(id)
        console.log(status)
        console.log(timestamp)
        console.log(parsedBody)

        const handler = this.handlers.get(id.toString());
        const rejecter = this.rejecters.get(id.toString());

        this.handlers.delete(id.toString());
        this.rejecters.delete(id.toString());

        const STATUS_OK = 0;
        if (status !== STATUS_OK) {
            const errMsg = typeof parsedBody === 'string'
                ? parsedBody
                : (parsedBody as any)?.error || (parsedBody as any)?.message || `Server error (status ${status})`;
            const error = new Error(errMsg) as any;
            error.status = status;
            if (rejecter) rejecter(error);
        } else {
            if (handler) handler({ id, status, timestamp, body: parsedBody });
        }
    }

    send(msgType: MsgType, payload: Uint8Array): void {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;

        if (msgType === MsgType.Heartbeat) {
            const buffer = new ArrayBuffer(HEADER_MIN_LEN);
            const uint8 = new Uint8Array(buffer);
            uint8.set(MAGIC_BYTES, 0);
            uint8[MAGIC_LEN] = msgType;
            this.ws.send(buffer);
            return;
        }

        const totalLen = HEADER_FULL_LEN + payload.length;
        const buffer = new ArrayBuffer(totalLen);
        const view = new DataView(buffer);
        const uint8 = new Uint8Array(buffer);

        uint8.set(MAGIC_BYTES, 0);
        view.setUint8(MAGIC_LEN, msgType);
        view.setUint32(MAGIC_LEN + TYPE_LEN, payload.length, false); // Big Endian in envelope
        uint8.set(payload, HEADER_FULL_LEN);

        this.ws.send(buffer);
    }

    private rejecters: Map<string, (err: any) => void> = new Map();

    request(path: string, body: any): Promise<any> {
        return new Promise((resolve, reject) => {
            if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
                return reject(new Error("Network error: WebSocket is not connected"));
            }

            const requestId = BigInt(Math.floor(Math.random() * 1000000));
            const pathBytes = new TextEncoder().encode(path);
            const bodyStr = typeof body === 'object' ? JSON.stringify(body) : String(body);
            const bodyBytes = new TextEncoder().encode(bodyStr);

            console.log("request", path, body)

            // 5 second timeout
            const timeout = setTimeout(() => {
                this.handlers.delete(requestId.toString());
                this.rejecters.delete(requestId.toString());
                reject(new Error("Request timed out: No response from server"));
            }, 5000);

            const wrappedStore = (data: any) => {
                clearTimeout(timeout);
                resolve(data);
            };

            const wrappedReject = (err: any) => {
                clearTimeout(timeout);
                reject(err);
            };

            // Request binary: Id(8) + PathLen(2) + Path + BodyLen(4) + Body (LittleEndian)
            const totalLen = 8 + 2 + pathBytes.length + 4 + bodyBytes.length;
            const buffer = new ArrayBuffer(totalLen);
            const view = new DataView(buffer);
            const uint8 = new Uint8Array(buffer);

            let offset = 0;
            view.setBigUint64(offset, requestId, true);
            offset += 8;

            view.setUint16(offset, pathBytes.length, true);
            offset += 2;
            uint8.set(pathBytes, offset);
            offset += pathBytes.length;

            view.setUint32(offset, bodyBytes.length, true);
            offset += 4;
            uint8.set(bodyBytes, offset);

            this.handlers.set(requestId.toString(), wrappedStore);
            this.rejecters.set(requestId.toString(), wrappedReject);
            this.send(MsgType.Request, uint8);
        });
    }

    onMessage(callback: MessageHandler) {
        this.onMessageCallback = callback;
    }

    close() {
        this.stopHeartbeat();
        this.ws?.close();
    }

    private startHeartbeat() {
        this.stopHeartbeat(); // Clear existing if any
        this.heartbeatInterval = setInterval(() => {
            console.log("Sending heartbeat...");
            this.send(MsgType.Heartbeat, new Uint8Array(0));
        }, 30000); // 30 seconds
    }

    private stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }
    }
}
