import React, { useState } from "react";
import { models } from "../../wailsjs/go/models";
import appIcon from "../assets/images/appicon.png";

interface SettingsProps {
  settings: models.Settings;
  onSettingsChange: (settings: models.Settings) => void;
  onClose: () => void;
  isVisible: boolean;
}

const Settings: React.FC<SettingsProps> = ({
  settings,
  onSettingsChange,
  onClose,
  isVisible,
}) => {
  const [activeTab, setActiveTab] = useState<"general" | "advanced" | "about">(
    "general"
  );
  const [localSettings, setLocalSettings] = useState<models.Settings>(settings);

  const handleSave = () => {
    onSettingsChange(localSettings);
    onClose();
  };

  const handleReset = () => {
    const defaultSettings = new models.Settings({
      id: settings.id,
      globalHotkey: "Cmd+Shift+Space",
      previousItemHotkey: "Cmd+Shift+C",
      pollingInterval: 500,
      maxItems: 100,
      maxDays: 7,
      autoLaunch: true,
      enableSounds: false,
      monitoringEnabled: true,
      allowPasswords: false,
      sortByRecent: "copied",
      createdAt: settings.createdAt,
      updatedAt: new Date(),
    });
    setLocalSettings(defaultSettings);
  };

  if (!isVisible) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/30 backdrop-blur-sm z-40 transition-opacity duration-200"
        onClick={onClose}
      />

      {/* Settings Window */}
      <div className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50 w-[520px] h-[650px] bg-macos-bg-primary dark:bg-macos-dark-bg-primary rounded-macos shadow-macos dark:shadow-macos-dark border border-macos-border dark:border-macos-dark-border transition-all duration-200 ease-out scale-100 opacity-100 flex flex-col">
        {/* Title Bar */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-macos-border dark:border-macos-dark-border flex-shrink-0">
          <h1 className="text-lg font-semibold text-macos-text-primary dark:text-macos-dark-text-primary">
            Klipd Preferences
          </h1>
          <button
            onClick={onClose}
            className="w-6 h-6 rounded-full bg-macos-accent-red hover:bg-red-500 flex items-center justify-center text-white text-xs transition-colors font-bold"
          >
            X
          </button>
        </div>

        {/* Tab Navigation */}
        <div className="flex border-b border-macos-border dark:border-macos-dark-border flex-shrink-0">
          {[
            { id: "general", label: "General", icon: "⚙️" },
            { id: "advanced", label: "Advanced", icon: "🔧" },
            { id: "about", label: "About", icon: "ℹ️" },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`flex-1 flex items-center justify-center px-4 py-3 text-sm font-medium transition-colors ${
                activeTab === tab.id
                  ? "text-macos-accent-blue dark:text-macos-dark-accent-blue border-b-2 border-macos-accent-blue dark:border-macos-dark-accent-blue"
                  : "text-macos-text-secondary dark:text-macos-dark-text-secondary hover:text-macos-text-primary dark:hover:text-macos-dark-text-primary"
              }`}
            >
              <span className="mr-2">{tab.icon}</span>
              {tab.label}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        <div className="flex-1 overflow-y-auto p-6 min-h-0">
          {activeTab === "general" && (
            <div className="space-y-6">
              {/* Global Hotkeys */}
              <div>
                <h3 className="text-base font-semibold text-macos-text-primary dark:text-macos-dark-text-primary mb-3">
                  Global Hotkeys
                </h3>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <label className="text-sm text-macos-text-primary dark:text-macos-dark-text-primary">
                      Show clipboard history
                    </label>
                    <input
                      type="text"
                      value={localSettings.globalHotkey}
                      onChange={(e) =>
                        setLocalSettings((prev) => {
                          const updated = new models.Settings({
                            ...prev,
                            globalHotkey: e.target.value,
                          });
                          return updated;
                        })
                      }
                      className="w-48 px-3 py-1 text-sm bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary border border-macos-border dark:border-macos-dark-border rounded-macos-input text-center"
                      placeholder="⌘⇧V"
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <label className="text-sm text-macos-text-primary dark:text-macos-dark-text-primary">
                      Copy Previous item to clipboard
                    </label>
                    <input
                      type="text"
                      value={localSettings.previousItemHotkey}
                      disabled={true}
                      onChange={(e) =>
                        setLocalSettings((prev) => {
                          const updated = new models.Settings({
                            ...prev,
                            previousItemHotkey: e.target.value,
                          });
                          return updated;
                        })
                      }
                      // disabled so greyed out
                      className={`w-48 px-3 py-1 text-sm bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary border border-macos-border dark:border-macos-dark-border rounded-macos-input text-center opacity-60 cursor-not-allowed`}
                      placeholder="⌘⇧C"
                    />
                  </div>
                </div>
              </div>

              {/* Clipboard Monitoring */}
              <div>
                <h3 className="text-base font-semibold text-macos-text-primary dark:text-macos-dark-text-primary mb-3">
                  Clipboard Monitoring
                </h3>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <label className="text-sm text-macos-text-primary dark:text-macos-dark-text-primary">
                        Polling interval
                      </label>
                      <p className="text-xs text-macos-text-secondary dark:text-macos-dark-text-secondary">
                        How often to check for clipboard changes
                      </p>
                    </div>
                    <select
                      value={localSettings.pollingInterval}
                      onChange={(e) =>
                        setLocalSettings((prev) => {
                          const updated = new models.Settings({
                            ...prev,
                            pollingInterval: Number(e.target.value),
                          });
                          return updated;
                        })
                      }
                      className="px-3 py-1 text-sm bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary border border-macos-border dark:border-macos-dark-border rounded-macos-input"
                    >
                      <option value={100}>100ms (High CPU)</option>
                      <option value={250}>250ms (Medium CPU)</option>
                      <option value={500}>500ms (Recommended)</option>
                      <option value={1000}>1000ms (Low CPU)</option>
                      <option value={2000}>2000ms (Very Low CPU)</option>
                    </select>
                  </div>
                </div>
              </div>

              {/* History Management */}
              <div>
                <h3 className="text-base font-semibold text-macos-text-primary dark:text-macos-dark-text-primary mb-3">
                  History Management
                </h3>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <label className="text-sm text-macos-text-primary dark:text-macos-dark-text-primary">
                        Maximum items
                      </label>
                      <p className="text-xs text-macos-text-secondary dark:text-macos-dark-text-secondary">
                        Auto-delete oldest items after this limit
                      </p>
                    </div>
                    <input
                      type="number"
                      min="10"
                      max="1000"
                      value={localSettings.maxItems}
                      onChange={(e) =>
                        setLocalSettings((prev) => {
                          const updated = new models.Settings({
                            ...prev,
                            maxItems: Number(e.target.value),
                          });
                          return updated;
                        })
                      }
                      className="w-20 px-3 py-1 text-sm bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary border border-macos-border dark:border-macos-dark-border rounded-macos-input text-center"
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <div>
                      <label className="text-sm text-macos-text-primary dark:text-macos-dark-text-primary">
                        Maximum days
                      </label>
                      <p className="text-xs text-macos-text-secondary dark:text-macos-dark-text-secondary">
                        Auto-delete items older than this
                      </p>
                    </div>
                    <input
                      type="number"
                      min="1"
                      max="365"
                      value={localSettings.maxDays}
                      onChange={(e) =>
                        setLocalSettings((prev) => {
                          const updated = new models.Settings({
                            ...prev,
                            maxDays: Number(e.target.value),
                          });
                          return updated;
                        })
                      }
                      className="w-20 px-3 py-1 text-sm bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary border border-macos-border dark:border-macos-dark-border rounded-macos-input text-center"
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <div>
                      <label className="text-sm text-macos-text-primary dark:text-macos-dark-text-primary">
                        Sort priority
                      </label>
                      <p className="text-xs text-macos-text-secondary dark:text-macos-dark-text-secondary">
                        After pinned items, show recently copied or recently
                        pasted first
                      </p>
                    </div>
                    <select
                      value={localSettings.sortByRecent || "copied"}
                      onChange={(e) =>
                        setLocalSettings((prev) => {
                          const updated = new models.Settings({
                            ...prev,
                            sortByRecent: e.target.value as "copied" | "pasted",
                          });
                          return updated;
                        })
                      }
                      className="px-3 py-1 text-sm bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary border border-macos-border dark:border-macos-dark-border rounded-macos-input"
                    >
                      <option value="copied">Recently copied</option>
                      <option value="pasted">Recently pasted</option>
                    </select>
                  </div>
                </div>
              </div>

              {/* System Integration */}
              <div>
                <h3 className="text-base font-semibold text-macos-text-primary dark:text-macos-dark-text-primary mb-3">
                  System Integration
                </h3>
                <div className="space-y-3">
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={localSettings.autoLaunch}
                      onChange={(e) =>
                        setLocalSettings((prev) => {
                          const updated = new models.Settings({
                            ...prev,
                            autoLaunch: e.target.checked,
                          });
                          return updated;
                        })
                      }
                      className="mr-3 rounded"
                    />
                    <div>
                      <span className="text-sm text-macos-text-primary dark:text-macos-dark-text-primary">
                        Launch at login
                      </span>
                      <p className="text-xs text-macos-text-secondary dark:text-macos-dark-text-secondary">
                        Start Klipd automatically when you log in
                      </p>
                    </div>
                  </label>
                  {/* <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={localSettings.enableSounds}
                      onChange={(e) =>
                        setLocalSettings((prev) => {
                          const updated = new models.Settings({
                            ...prev,
                            enableSounds: e.target.checked,
                          });
                          return updated;
                        })
                      }
                      className="mr-3 rounded"
                    />
                    <div>
                      <span className="text-sm text-macos-text-primary dark:text-macos-dark-text-primary">
                        Enable sounds
                      </span>
                      <p className="text-xs text-macos-text-secondary dark:text-macos-dark-text-secondary">
                        Play sound when copying large items
                      </p>
                    </div>
                  </label> */}
                </div>
              </div>
            </div>
          )}

          {activeTab === "advanced" && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium text-macos-text-primary dark:text-macos-dark-text-primary mb-4">
                  Privacy & Security
                </h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <label className="text-sm font-medium text-macos-text-primary dark:text-macos-dark-text-primary">
                        Allow Password Capture
                      </label>
                      <p className="text-xs text-macos-text-secondary dark:text-macos-dark-text-secondary mt-1">
                        When enabled, passwords and sensitive content will be
                        captured in clipboard history
                      </p>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.allowPasswords}
                        onChange={(e) =>
                          setLocalSettings((prev) => {
                            const updated = new models.Settings({
                              ...prev,
                              allowPasswords: e.target.checked,
                            });
                            return updated;
                          })
                        }
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-macos-bg-tertiary dark:bg-macos-dark-bg-tertiary peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-macos-accent-blue"></div>
                    </label>
                  </div>
                </div>
              </div>

              <div>
                <h3 className="text-lg font-medium text-macos-text-primary dark:text-macos-dark-text-primary mb-4">
                  Monitoring
                </h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <label className="text-sm font-medium text-macos-text-primary dark:text-macos-dark-text-primary">
                        Clipboard Monitoring
                      </label>
                      <p className="text-xs text-macos-text-secondary dark:text-macos-dark-text-secondary mt-1">
                        Enable or disable clipboard content monitoring
                      </p>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.monitoringEnabled}
                        onChange={(e) =>
                          setLocalSettings((prev) => {
                            const updated = new models.Settings({
                              ...prev,
                              monitoringEnabled: e.target.checked,
                            });
                            return updated;
                          })
                        }
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-macos-bg-tertiary dark:bg-macos-dark-bg-tertiary peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-macos-accent-blue"></div>
                    </label>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === "about" && (
            <div className="space-y-6">
              <div className="text-center py-8">
                <img
                  src={appIcon}
                  alt="Klipd"
                  className="w-16 h-16 mx-auto mb-4"
                />
                <h2 className="text-2xl font-bold text-macos-text-primary dark:text-macos-dark-text-primary mb-2">
                  Klipd
                </h2>
                <p className="text-sm text-macos-text-secondary dark:text-macos-dark-text-secondary mb-4">
                  Version 1.0.0
                </p>
                <p className="text-sm text-macos-text-secondary dark:text-macos-dark-text-secondary max-w-sm mx-auto leading-relaxed">
                  A fast, intelligent clipboard manager for macOS. Built with
                  Wails, Go, and React.
                </p>
              </div>

              <div className="space-y-3">
                {/* <button className="w-full flex items-center justify-center px-4 py-2 text-sm bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary border border-macos-border dark:border-macos-dark-border rounded-macos-button hover:bg-macos-bg-tertiary dark:hover:bg-macos-dark-bg-tertiary transition-colors">
                  <span className="mr-2">📄</span>
                  Export Clipboard History
                </button> */}
                <button
                  onClick={handleReset}
                  className="w-full flex items-center justify-center px-4 py-2 text-sm bg-macos-accent-red text-white rounded-macos-button hover:bg-red-500 transition-colors"
                >
                  <span className="mr-2">↻</span>
                  Reset All Settings
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Footer Buttons */}
        <div className="flex items-center justify-end space-x-3 px-6 py-4 border-t border-macos-border dark:border-macos-dark-border flex-shrink-0 bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary rounded-b-macos">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-macos-text-primary dark:text-macos-dark-text-primary hover:bg-macos-bg-tertiary dark:hover:bg-macos-dark-bg-tertiary rounded-macos-button transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            className="px-4 py-2 text-sm bg-macos-accent-blue text-white rounded-macos-button hover:bg-blue-500 transition-colors"
          >
            Save Changes
          </button>
        </div>
      </div>
    </>
  );
};

export default Settings;
