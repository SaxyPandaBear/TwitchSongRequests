import { Component, OnInit } from "@angular/core";
import { NgWizardConfig, THEME, NgWizardService } from "ng-wizard";
import { ActivatedRoute } from "@angular/router";
import { SpotifyService } from "../../spotify.service";
import { OauthService } from "../../oauth.service";
@Component({
  selector: "app-login",
  templateUrl: "./login.component.html",
  styleUrls: ["./login.component.scss"],
})
export class LoginComponent implements OnInit {
  constructor(
    private route: ActivatedRoute,
    private ngWizardService: NgWizardService,
    private spotifyService: SpotifyService,
    private oauthService: OauthService
  ) {
    this.route.fragment.subscribe((fragment: string) => {
      // TODO: include robust logic to check for access code -- for now, we'll default to simple impl
      if (fragment && fragment.includes("access_token")) {
        const access_token = window.location.href.match(
          /\#(?:access_token)\=([\S\s]*?)\&/
        )[1];
        this.isLoading = true;

        //simulate api response of one second
        setTimeout(() => {
          if (!this.twitchAccessToken) {
            this.setTwitchAccessToken(access_token);
          } else if (!this.spotifyAccessToken) {
            this.setSpotifyAccessToken(access_token);
          }
          this.isLoading = false;
          this.ngWizardService.next();
        }, 1000);

        //TODO: invoke service to pass to api, await api response before allowing user to continue
      }
    });
  }
  config: NgWizardConfig;

  ngOnInit(): void {
    this.oauthService.getOauthStatus().subscribe((currentUsersAccessTokens) => {
      const {
        spotifyAccessKey,
        twitchAccessKey,
      } = currentUsersAccessTokens as any;
      console.log({ currentUsersAccessTokens });
      this.spotifyAccessToken = spotifyAccessKey;
      this.twitchAccessToken = twitchAccessKey;
      if (twitchAccessKey) {
        this.ngWizardService.next();
      }
      if (spotifyAccessKey) {
        this.ngWizardService.next();
      }

      this.config = {
        selected: this.getSelectedStep(),
        theme: THEME.arrows,
        toolbarSettings: {
          toolbarExtraButtons: [
            {
              text: "Finish",
              class: "btn btn-info",
              event: () => {
                alert("Finished!!!");
              },
            },
          ],
        },
      };
    });
    // this.checkSessionStorageAndAssignValues();
  }
  stepChanged(event) {
    console.log({ event });
  }

  setSpotifyAccessToken(token) {
    console.log("setting spotify access token");

    //this.spotifyAccessToken = token;
    //sessionStorage.setItem("spotifyAccessToken", "true");
    this.oauthService.setSpotifyAcessKey(token).subscribe(console.log);
  }
  setTwitchAccessToken(token) {
    console.log("setting twtich access token");
    //  this.twitchAccessToken = token;
    //sessionStorage.setItem("twitchAccessToken", "true");
    this.oauthService.setTwitchAccessKey(token).subscribe(console.log);
  }
  checkSessionStorageAndAssignValues() {
    const potentialSpotifyToken = sessionStorage.getItem("spotifyAccessToken");
    const potentialTwitchToken = sessionStorage.getItem("twitchAccessToken");

    if (potentialSpotifyToken || potentialTwitchToken) {
      if (potentialSpotifyToken) {
        this.setSpotifyAccessToken(potentialSpotifyToken);
      }
      if (potentialTwitchToken) {
        this.setTwitchAccessToken(potentialTwitchToken);
      }
    }
  }

  getUserDevices() {
    this.spotifyService
      .getPlayer(this.spotifyAccessToken)
      .subscribe(
        ({
          device: { is_active, name: deviceName },
          item: { name: songName },
        }) => {
          console.log({ deviceName, is_active, songName });
          if (is_active) {
            this.currentSong = songName;
            this.playingOn = deviceName;
            this.playerIsActive = is_active;
          }
        }
      );
  }

  getSelectedStep() {
    if (this.twitchAccessToken && this.spotifyAccessToken) {
      return 2;
    } else if (this.twitchAccessToken) {
      return 1;
    }
    return 0;
  }
  isLoading = false;

  twitchAccessToken = undefined;
  spotifyScope = "user-modify-playback-state%20user-read-playback-state";

  localPath = "http%3A%2F%2Flocalhost%3A4200";
  spotifyClientId = "5b0a6304d93b4f2b9c6bbf27e7db5592";
  redirectPathTwo = `https://id.twitch.tv/oauth2/authorize?client_id=n43pmbmxpn1xgtd36oraj6y4xxpp2h&redirect_uri=${this.localPath}&response_type=token&scope=channel%3Aread%3Aredemptions`;

  spotifyRedirectUri = `https://accounts.spotify.com/authorize?client_id=${this.spotifyClientId}&redirect_uri=${this.localPath}&response_type=token&scope=${this.spotifyScope}`;
  spotifyAccessToken = undefined;

  currentSong;
  playingOn;
  playerIsActive;
}
