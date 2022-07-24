import { useEffect, useRef, useState } from "react";
import TextareaAutosize from "react-textarea-autosize";
import { validate } from "uuid";
import { equals, Talkog } from "./talkog";
import { useInterval } from "./useInterval";

export const App = () => {
  const [userId, setUserId] = useState("");
  const [talkogs, setTalkogs] = useState<Talkog[]>([]);
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    socketRef.current = new WebSocket(
      import.meta.env.VITE_WEBSOCKET_SERVER_URL
    );

    socketRef.current.addEventListener("message", ({ data }) => {
      const json = JSON.parse(data);
      const talkog: Talkog = {
        userId: json.UserId,
        talk: json.Talk,
        timestamp: new Date(parseInt(json.Timestamp) * 1000), // Unix Time -> Date type
      };
      setTalkogs((talkogs) => {
        if (talkogs.some((v) => equals(v, talkog))) {
          return talkogs;
        } else {
          return [talkog, ...talkogs];
        }
      });
    });

    return () => {
      socketRef.current?.close();
    };
  }, []);

  useEffect(() => {
    if (!validate(userId)) {
      setTalkogs((_) => []);
    }
  }, [userId]);

  useInterval(() => {
    if (validate(userId)) {
      socketRef.current?.send(
        JSON.stringify({
          action: "fetchTalkogs",
          UserId: userId,
        })
      );
    }
  }, 1000);

  return (
    <div className="container mx-auto my-20">
      <TextareaAutosize
        className={
          "textarea resize-none flex m-auto text-center p-4 text-5xl" +
          (userId === "" ? " " + "textarea-primary" : " " + "textarea-ghost")
        }
        placeholder="Type your user ID here"
        onChange={(e) => setUserId(e.target.value)}
      />

      <div className="mx-40 my-28">
        {talkogs.length === 0 ? (
          <></>
        ) : (
          talkogs
            .map(({ talk, timestamp }) => (
              <div className="grid gap-20 grid-cols-4">
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
