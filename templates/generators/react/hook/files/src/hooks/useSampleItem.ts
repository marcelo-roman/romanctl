import { useState } from "react";

export default function useSampleItem() {
  const [value, setValue] = useState(0);
  return { value, setValue };
}
