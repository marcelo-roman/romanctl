import { createContext, useContext } from "react";

export interface JobTracker {
  tracked: string[];
  track: (id: string) => void;
  untrack: (id: string) => void;
}

export const JobTrackerContext = createContext<JobTracker>({
  tracked: [],
  track: () => undefined,
  untrack: () => undefined,
});

export function useJobTracker(): JobTracker {
  return useContext(JobTrackerContext);
}
