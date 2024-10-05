import {configDir} from "@tauri-apps/api/path";
import {createDir, readTextFile, writeFile} from "@tauri-apps/api/fs";

const settingsDirectoryName = "tech.tanay.freemind";
const settingsFileName = "settings.json";

type Settings = {
  timerValue: number,
  websiteBlockList: Record<string, string[]>
}

const defaultSettings: Settings = {
  timerValue: 25,
  websiteBlockList: {
    "Social Media": ["facebook.com", "twitter.com", "instagram.com"],
    "Shopping": ["amazon.com", "flipkart.com"],
    "Work": ["slack.com", "github.com"],
    "News": ["nytimes.com", "cnn.com"],
    "Entertainment": ["netflix.com", "youtube.com"],
    "Games": ["steam.com", "epicgames.com"],
  }
}

const mergeSettings = (oldSettings: Settings, newSettings: Settings): Settings => {
  console.log(`before merging settings ${JSON.stringify(oldSettings)}`)
  console.log(`before merging settings ${JSON.stringify(newSettings)}`)
  const newVar = {...oldSettings.websiteBlockList, ...newSettings.websiteBlockList};
  console.log(`merging settings ${JSON.stringify(oldSettings)}`)
  console.log(`merging settings ${JSON.stringify(newSettings)}`)
  console.log(`merging settings ${JSON.stringify(newVar)}`)
  return {
    timerValue: newSettings.timerValue,
    websiteBlockList: newVar,
  }
}

const saveSettings = async (settings: Settings) => {
  console.log(`saving settings`)
  const configBase = await configDir();
  const settingsStem = `${configBase}${settingsDirectoryName}`;
  await createDir(settingsStem, {recursive: true});
  const settingsPath = `${configBase}${settingsDirectoryName}/${settingsFileName}`;
  const newSettings = mergeSettings(await loadSettings(), settings);
  console.log(`saving settings ${JSON.stringify(newSettings)}`)
  await writeFile(settingsPath, JSON.stringify(newSettings));
}

const loadSettings = async (): Promise<Settings> => {
  console.log(`loading settings`)
  const configBase = await configDir();
  const settingsPath = `${configBase}${settingsDirectoryName}/${settingsFileName}`;
  try {
    await readTextFile(settingsPath);
  }
  catch (e) {
    await saveSettings(defaultSettings);
    return defaultSettings;
  }
  const existingSettings: Settings = JSON.parse(await readTextFile(settingsPath));
  console.log(`Default settings ${JSON.stringify(defaultSettings)}`)
  const settings = mergeSettings(defaultSettings, existingSettings);
  console.log(`loaded settings ${JSON.stringify(settings)}`)
  return settings
}

export {
  saveSettings,
  loadSettings,
}
