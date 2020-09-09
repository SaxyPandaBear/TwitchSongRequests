import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";

@Injectable({
  providedIn: "root",
})
export class OauthService {
  endpoint = "http://localhost:8080/api/session";
  spotifyEndpoint = `${this.endpoint}/spotify`;
  twitchcEndpoint = `${this.endpoint}/twitch`;
  accessKeysEndpoint = `${this.endpoint}/access-keys`;

  constructor(private http: HttpClient) {}

  getOauthStatus() {
    return this.http.get(this.accessKeysEndpoint, { withCredentials: true });
  }
  setSpotifyAcessKey(spotifyAccessKey: string) {
    return this.http.post(
      this.spotifyEndpoint,
      {
        accessKey: spotifyAccessKey,
      },
      { withCredentials: true }
    );
  }
  // setTwitchAccessKey(twitchAccesKey: string, channelId) {
  //   return this.http.post(
  //     this.twitchcEndpoint,
  //     { accessKeysAndChannelId: { accessKey: twitchAccesKey, channelId } },
  //     { withCredentials: true }
  //   );
  // }

  setTwitchAccessKey(twitchCode: string) {
    return this.http.post(
      this.twitchcEndpoint,
      { accessKey: twitchCode },
      { withCredentials: true }
    );
  }
}
