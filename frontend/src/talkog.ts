export type Talkog = {
  userId: string;
  talk: string;
  timestamp: Date;
};

export const equals = (t1: Talkog, t2: Talkog) => {
  //   console.log(t1.userId, t1.talk, t1.timestamp);
  //   console.log(t2.userId, t2.talk, t2.timestamp);
  //   console.log("");
  return (
    t1.userId === t2.userId &&
    t1.talk === t2.talk &&
    t1.timestamp.getTime() === t2.timestamp.getTime()
  );
};
