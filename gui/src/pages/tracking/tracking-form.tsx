import React from "react";
import { useTranslation } from "react-i18next";
import { motion } from "framer-motion";

import { CFNMachineContext } from "@/main/machine";
import { GetAvailableLogs } from "@@/go/core/CommandHandler";

import { ActionButton } from "@/ui/action-button";
import { Checkbox } from "@/ui/checkbox";
import { data } from "@@/go/models";
import { PageHeader } from "@/ui/page-header";

export const TrackingForm: React.FC = () => {
  const { t } = useTranslation();
  const [_, send] = CFNMachineContext.useActor();

  const playerIdInputRef = React.useRef<HTMLInputElement>(null);
  const restoreRef = React.useRef<HTMLInputElement>(null);

  const [playerIdInput, setPlayerIdInput] = React.useState<string>("");
  const [oldPlayers, setOldPlayers] = React.useState<data.PlayerInfo[] | null>(
    null
  );

  React.useEffect(() => {
    GetAvailableLogs().then((logs) => setOldPlayers(logs));
  }, []);

  const onSubmit: React.FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    if (playerIdInput == "") return;
    send({
      type: "submit",
      playerInfo: {
        displayName: oldPlayers?.find((old) => old.id == playerIdInput) ?? playerIdInput,
        id: playerIdInput,
      },
      restore: restoreRef.current && restoreRef.current.checked,
    });
  };

  const playerChipClicked = (player: data.PlayerInfo) => {
    if (playerIdInputRef.current) {
      playerIdInputRef.current.value = player.id;
      setPlayerIdInput(player.id);
    }
  };

  return (
    <>
      <PageHeader text={t("startTracking")} />
      <motion.form
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.125 }}
        className="relative h-full overflow-y-scroll overflow-x-visible w-full justify-self-center flex flex-col pt-12 gap-5 px-40 pb-4"
        onSubmit={onSubmit}
      >
        <h3 className="text-lg">{t("enterCfnName")}</h3>
        <input
          ref={playerIdInputRef}
          onChange={(e) => setPlayerIdInput(e.target.value)}
          className="bg-transparent border-b-2 border-0 focus:ring-offset-transparent focus:ring-transparent border-b-[rgba(255,255,255,0.275)] focus:border-white hover:border-white outline-none focus:outline-none hover:text-white transition-colors py-3 px-4 block w-full text-lg text-gray-300"
          type="text"
          placeholder={t("cfnName")!}
          autoCapitalize="off"
          autoComplete="off"
          autoCorrect="off"
          autoSave="off"
        />
        {oldPlayers && (
          <div className="flex flex-wrap gap-2 content-center items-center text-center">
            {oldPlayers.map((player) => (
              <button
                key={player.displayName}
                type="button"
                onClick={() => playerChipClicked(player)}
                className="whitespace-nowrap bg-[rgb(255,255,255,0.075)] hover:bg-[rgb(255,255,255,0.125)] text-base rounded-2xl transition-all items-center px-3 pt-1"
              >
                {player.displayName}
              </button>
            ))}
          </div>
        )}
        <footer className="flex items-center w-full">
          {oldPlayers && oldPlayers.some(old => old.id == playerIdInput) && (
            <div className="group flex items-center">
              <Checkbox ref={restoreRef} id="restore-session" />
              <label htmlFor="restore-session" className="text-lg cursor-pointer text-gray-300 group-hover:text-white transition-colors">
                {t("restoreSession")}
              </label>
            </div>
          )}
          <ActionButton type="submit" style={{ filter: "hue-rotate(-65deg)" }} className="ml-auto">
            {t("start")}
          </ActionButton>
        </footer>
      </motion.form>
    </>
  );
};
