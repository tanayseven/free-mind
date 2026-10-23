# NixOS module. /etc/systemd/system is read-only on NixOS, so this runs the
# daemon instead of the app's own install, and writes the marker that tells
# the app the daemon is package-managed (install::MANAGED_MARKER).
self:
{
  config,
  lib,
  pkgs,
  ...
}:

let
  cfg = config.services.free-mind;
in
{
  options.services.free-mind = {
    enable = lib.mkEnableOption "Free Mind and its website blocking daemon";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      description = "The Free Mind package to use.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      example = "alice";
      description = "The user allowed to start and stop blocking.";
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    environment.etc."free-mind/managed-daemon".text = "";

    # The daemon edits /etc/hosts, which NixOS otherwise links read-only from
    # the store. A copied file keeps the networking.hosts entries.
    environment.etc.hosts.mode = "0644";

    systemd.services.free-mind-daemon = {
      description = "Free Mind website blocking helper";
      wantedBy = [ "multi-user.target" ];
      script = ''
        exec ${lib.getExe cfg.package} --daemon --owner-uid "$(id -u ${lib.escapeShellArg cfg.user})"
      '';
      serviceConfig.Restart = "always";
    };
  };
}
