import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

interface CFNTrackerTranslation {
  tracking: string,
  history: string,
  language: string,
  startTracking: string,
  cfnName: string,
  start: string,
  opponent: string,
  character: string,
  lpGain: string,
  delete: string,
  goBack: string,
  loading: string,
  wins: string,
  losses: string,
  winRate: string,
  stop: string,
  openResultFolder: string,
  enterCfnName: string,
  result: string,
  time: string,
  winStreak: string,
  newVersionAvailable: string,
  pause: string,
  unpause: string,
  statistics: string,
  date: string,
  minimize: string,
  restoreSession: string,
  export: string,
  league: string
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  // init i18next
  // for all options read: https://www.i18next.com/overview/configuration-options
  .init({
    debug: true,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false,
    },
    resources: {
      en: {
        translation: <CFNTrackerTranslation>{
          tracking: 'Tracking',
          history: 'Match Log',
          language: 'Language',
          startTracking: 'Start Tracking',
          cfnName: 'CFN Name',
          start: 'Start',
          opponent: 'Opponent',
          character: 'Character',
          lpGain: 'LP Gain',
          delete: 'Clear Log',
          goBack: 'Go back',
          loading: 'Loading, please wait',
          wins: 'Wins',
          losses: 'Losses',
          winRate: 'Win Rate',
          stop: 'Stop',
          openResultFolder: 'Results Folder',
          enterCfnName: 'Enter your CFN name',
          result: 'Result',
          time: 'Time',
          winStreak: 'Win Streak',
          newVersionAvailable: 'Update available!',
          pause: 'Pause',
          unpause: 'Unpause',
          statistics: 'Statistics',
          date: 'Date',
          minimize: 'minimize',
          restoreSession: 'Restore last session',
          export: 'Export CSV',
          league: 'League'
        }
      },
      fr: {
        translation: <CFNTrackerTranslation>{
          tracking: 'Suivie',
          history: 'Histoire',
          language: 'Langue',
          startTracking: 'Démarrer le suivi',
          cfnName: 'Nom CFN',
          start: 'Début',
          opponent: 'Ennemi',
          character: 'Caractère',
          lpGain: 'Gain LP',
          delete: 'Supprimer le journal',
          goBack: 'Retourner',
          loading: 'Chargement, veuillez patienter',
          wins: 'Gagne',
          losses: 'Pertes',
          winRate: 'Taux de réussite',
          stop: 'Arrêter',
          openResultFolder: 'Dossier des résultats',
          enterCfnName: 'Entrez votre nom CFN',
          result: 'Dossier',
          time: 'Temps',
          winStreak: 'Série de victoires',
          newVersionAvailable: 'Nouvelle version disponible!',
          pause: 'Pause',
          unpause: 'Reprendre',
          statistics: 'Statistiques',
          date: 'Date',
          minimize: 'réduire',
          restoreSession: 'Restaurer la dernière session',
          export: 'Exporter CSV',
          league: 'Ligue'
        }
      },
      jp: {
        translation: <CFNTrackerTranslation>{
          tracking: '追跡',
          history: 'ログ',
          language: '言語',
          startTracking: '追跡を開始',
          cfnName: 'CFN名',
          start: '始める',
          opponent: '敵',
          character: 'キャラクター',
          lpGain: 'LP得',
          delete: '日誌を削除',
          goBack: '戻る',
          loading: 'お待ちください',
          wins: '勝つ',
          losses: '負け',
          winRate: '勝率',
          stop: 'やめる',
          openResultFolder: '結果フォルダ',
          enterCfnName: 'CFN名を入力してください',
          result: '結果',
          time: '時間',
          winStreak: '連勝',
          newVersionAvailable: 'アップデートが入手可能!',
          pause: '休止',
          unpause: '続く',
          statistics: '統計',
          date: '日付',
          minimize: '最小化',
          restoreSession: '最後のセッションを復元',
          export: 'CSVに保存',
          league: 'リーグ'
        }
      }
    }
  });

export default i18n;