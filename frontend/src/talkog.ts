export type Talkog = {
  userId: string;
  talk: string;
  timestamp: Date;
};

export const create = (
  userId: string,
  talk: string,
  timestamp: number
): Talkog => {
  return {
    userId,
    talk,
    timestamp: new Date(timestamp * 1000),
  };
};

export const equals = (t1: Talkog, t2: Talkog) => {
  return (
    t1.userId === t2.userId &&
    t1.talk === t2.talk &&
    t1.timestamp.getTime() === t2.timestamp.getTime()
  );
};
