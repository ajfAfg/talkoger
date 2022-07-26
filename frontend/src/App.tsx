import { useEffect, useRef, useState } from "react";
import TextareaAutosize from "react-textarea-autosize";
import { validate } from "uuid";
import { create, Talkog } from "./talkog";

export const App = () => {
  const [userId, setUserId] = useState("");
  const [talkogs, setTalkogs] = useState<Talkog[]>([]);
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (validate(userId)) {
      socketRef.current = new WebSocket(
        import.meta.env.VITE_WEBSOCKET_SERVER_URL
      );

      socketRef.current.addEventListener("message", ({ data }) => {
        const { UserId, Talk, Timestamp } = JSON.parse(data);
        const talkog = create(UserId, Talk, parseInt(Timestamp));
        setTalkogs((talkogs) => {
          if (talkogs.length === 0) {
            return [talkog];
          }

          if (talkog.timestamp.getTime() < talkogs[0].timestamp.getTime()) {
            // Sort the data to be displayed in the correct order,
            // since new data may be received while the initial data is being received.
            return [talkog, ...talkogs]
              .sort(
                ({ timestamp: t1 }, { timestamp: t2 }) =>
                  t1.getTime() - t2.getTime()
              )
              .reverse();
          } else {
            return [talkog, ...talkogs];
          }
        });
      });

      socketRef.current.addEventListener("open", (_) => {
        socketRef.current?.send(
          JSON.stringify({
            action: "fetchTalkogs",
            UserId: userId,
          })
        );
      });

      return () => {
        socketRef.current?.close();
      };
    } else {
      setTalkogs((_) => []);
    }
  }, [userId]);

  return (
    <div className="container mx-auto my-20">
      <TextareaAutosize
        className={
          "textarea resize-none flex m-auto text-center p-4 text-5xl" +
          (userId === "" ? " " + "textarea-primary" : " " + "textarea-ghost")
        }
        placeholder="Type your user ID here"
        onChange={(e) => setUserId(e.target.value)}
        autoFocus
      />

      <div className="mx-40 my-28">
        {talkogs.length === 0 ? (
          <></>
        ) : (
          talkogs
            .map(({ talk, timestamp }, i) => (
              <div className="grid gap-20 grid-cols-4" key={i}>
                <div className="col-span-3">
                  <p className="break-words text-xl p-4">{talk}</p>
                </div>

                <div className="flex justify-center">
                  <div className="m-auto text-center">
                    {timestamp
                      .toLocaleString()
                      .split(", ")
                      .map((str) => (
                        <p>{str}</p>
                      ))}
                  </div>
                </div>
              </div>
            ))
            .reduce((acc, elem) => (
              <>
                {acc} <div className="divider" /> {elem}
              </>
            ))
        )}
      </div>
    </div>
  );
};
