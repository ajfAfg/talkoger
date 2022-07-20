import { useEffect } from "react";

export const useInterval = (callback: Function, delay?: number) => {
  useEffect(() => {
    const interval = setInterval(() => callback(), delay || 0);
    return () => clearInterval(interval);
  }, [callback, delay]);
};
