import React from "react";
import { useUiLocaleNamespace, useUiLocaleNamespacesActivation } from "../localization";

interface ErrorBoundaryProps {
  children: React.ReactNode;
  onReset?: () => void;
}

interface ErrorBoundaryState {
  hasError: boolean;
  errorMessage?: string;
}

interface ErrorBoundaryText {
  title: string;
  unknown: string;
  reload: string;
  exit: string;
}

class ErrorBoundaryImpl extends React.Component<
  ErrorBoundaryProps & { text: ErrorBoundaryText },
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, errorMessage: error?.message ?? "Unknown error" };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    if (import.meta.env.DEV) {
      console.error("[ErrorBoundary]", error, errorInfo);
    }
  }

  private handleReload = () => {
    window.location.reload();
  };

  private handleHome = () => {
    this.props.onReset?.();
    window.location.assign("/");
  };

  render() {
    if (!this.state.hasError) {
      return this.props.children;
    }

    return (
      <div className="fatalErrorScreen" role="alert">
        <div className="fatalErrorCard">
          <h2>{this.props.text.title}</h2>
          <p className="muted">{this.state.errorMessage ?? this.props.text.unknown}</p>
          <div className="fatalErrorActions">
            <button className="primary" onClick={this.handleReload}>
              {this.props.text.reload}
            </button>
            <button className="ghost" onClick={this.handleHome}>
              {this.props.text.exit}
            </button>
          </div>
        </div>
      </div>
    );
  }
}

export default function ErrorBoundary(props: ErrorBoundaryProps) {
  useUiLocaleNamespacesActivation(["reconnect", "misc", "common"]);
  const locale = useUiLocaleNamespace("reconnect", { fallbacks: ["misc", "common"] });
  const text: ErrorBoundaryText = {
    title: locale.t("fatalErrorTitle"),
    unknown: locale.t("fatalErrorUnknown"),
    reload: locale.t("fatalErrorReload"),
    exit: locale.t("errorScreenExitToMenu"),
  };

  return <ErrorBoundaryImpl {...props} text={text} />;
}
