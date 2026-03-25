import { useUiLocaleNamespace, useUiLocaleNamespacesActivation } from "../localization";

interface ErrorScreenProps {
  message: string;
  canRetry: boolean;
  reconnecting?: boolean;
  onRetry: () => void;
  onExitToMenu: () => void;
}

export default function ErrorScreen({
  message,
  canRetry,
  reconnecting = false,
  onRetry,
  onExitToMenu,
}: ErrorScreenProps) {
  useUiLocaleNamespacesActivation(["reconnect", "misc", "common"]);
  const text = useUiLocaleNamespace("reconnect", { fallbacks: ["misc", "common"] });

  return (
    <div className="errorScreen" role="alert">
      <div className="errorScreenCard">
        <h3>{text.t("errorScreenTitle")}</h3>
        <div>{message}</div>
        {reconnecting ? <div className="muted">{text.t("errorScreenReconnecting")}</div> : null}
        <div className="errorScreenActions">
          {canRetry ? (
            <button className="primary" onClick={onRetry}>
              {text.t("retryButton")}
            </button>
          ) : null}
          <button className="ghost" onClick={onExitToMenu}>
            {text.t("errorScreenExitToMenu")}
          </button>
        </div>
      </div>
    </div>
  );
}
