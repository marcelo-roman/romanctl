interface ErrorMessageProps {
  message?: string;
}

export function ErrorMessage({ message }: ErrorMessageProps) {
  if (!message) {
    return null;
  }
  return (
    <p role="alert" className="error">
      {message}
    </p>
  );
}
