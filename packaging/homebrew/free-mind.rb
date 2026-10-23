# Homebrew cask, for a tap (e.g. tanayseven/homebrew-tap) or homebrew/cask.
# The release workflow fills in {{VERSION}} and {{SHA256_DMG_UNIVERSAL}}.
cask "free-mind" do
  version "{{VERSION}}"
  sha256 "{{SHA256_DMG_UNIVERSAL}}"

  url "https://github.com/tanayseven/free-mind/releases/download/v#{version}/Free.Mind_#{version}_universal.dmg"
  name "Free Mind"
  desc "Block distracting websites and stay focused"
  homepage "https://github.com/tanayseven/free-mind"

  depends_on macos: ">= :catalina"

  app "Free Mind.app"

  # The daemon the app installs on first launch.
  uninstall launchctl: "tech.tanay.free-mind.daemon",
            delete:    [
              "/Library/LaunchDaemons/tech.tanay.free-mind.daemon.plist",
              "/Library/PrivilegedHelperTools/tech.tanay.free-mind.daemon",
            ]

  zap trash: [
    "~/Library/Caches/tech.tanay.free-mind",
    "~/Library/WebKit/tech.tanay.free-mind",
  ]
end
