

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

const MAGIC = "WUKONG";
const MAGIC_BYTES = new TextEncoder().encode(MAGIC);

export type MessageHandler = (msgType: MsgType, data: Uint8Array) => void;

export class WkClient {
    private ws: WebSocket | null = null;
    private handlers: Map<string, (data: any) => void> = new Map();
    private onMessageCallback: MessageHandler | null = null;

    constructor(private url: string) {}

    connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.ws = new WebSocket(this.url);
            this.ws.binaryType = "arraybuffer";

            this.ws.onopen = () => {
                console.log("Connected to WkServer");
                resolve();
            };

            this.ws.onerror = (err) => {
                console.error("WebSocket error:", err);
                reject(err);
            };

            this.ws.onmessage = (event) => {
                console.log("websocket on data")
                this.handleData(event.data);
            };

            this.ws.onclose = () => {
                console.log("Disconnected from WkServer");
            };
        });
    }

    private handleData(buffer: ArrayBuffer) {
        const uint8 = new Uint8Array(buffer);
        const view = new DataView(buffer);
        
        if (uint8.length < 7) return; 

        // Check MAGIC
        for (let i = 0; i < 6; i++) {
            if (uint8[i] !== MAGIC_BYTES[i]) return;
        }

        const msgType = view.getUint8(6) as MsgType;
        
        if (msgType === MsgType.Heartbeat) {
            if (this.onMessageCallback) this.onMessageCallback(msgType, new Uint8Array(0));
            return;
        }

        if (uint8.length < 11) return;
        const dataLen = view.getUint32(7, false); // Big Endian in protocol envelope
        const data = uint8.slice(11, 11 + dataLen);

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
        } catch(e) {}
        console.log(id)
        console.log(status)
        console.log(timestamp)
        console.log(parsedBody)

        const handler = this.handlers.get(id.toString());
        if (handler) {
            handler({ id, status, timestamp, body: parsedBody });
            this.handlers.delete(id.toString());
        }
    }

    send(msgType: MsgType, payload: Uint8Array): void {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;

        if (msgType === MsgType.Heartbeat) {
            const buffer = new ArrayBuffer(7);
            const uint8 = new Uint8Array(buffer);
            uint8.set(MAGIC_BYTES, 0);
            uint8[6] = msgType;
            this.ws.send(buffer);
            return;
        }

        const totalLen = 6 + 1 + 4 + payload.length;
        const buffer = new ArrayBuffer(totalLen);
        const view = new DataView(buffer);
        const uint8 = new Uint8Array(buffer);

        uint8.set(MAGIC_BYTES, 0);
        view.setUint8(6, msgType);
        view.setUint32(7, payload.length, false); // Big Endian in envelope
        uint8.set(payload, 11);

        this.ws.send(buffer);
    }

    request(path: string, body: any): Promise<any> {
        return new Promise((resolve) => {
            const requestId = BigInt(Math.floor(Math.random() * 1000000));
            const pathBytes = new TextEncoder().encode(path);
            const bodyStr = typeof body === 'object' ? JSON.stringify(body) : String(body);
            const bodyBytes = new TextEncoder().encode(bodyStr);

            console.log("request",path,body)

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

            this.handlers.set(requestId.toString(), resolve);
            this.send(MsgType.Request, uint8);
        });
    }

    onMessage(callback: MessageHandler) {
        this.onMessageCallback = callback;
    }

    close() {
        this.ws?.close();
    }
}
