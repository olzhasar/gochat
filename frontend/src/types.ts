export enum MessageType {
  TEXT = 1,
  NAME = 2,
  LEAVE = 3,
  TYPING = 4,
  STOP_TYPING = 5,
}

export interface Message {
  msgType: number;
  content: string;
  author: string | null;
}
