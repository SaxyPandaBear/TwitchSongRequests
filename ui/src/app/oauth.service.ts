import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Injectable({
    providedIn: 'root',
})
export class OauthService {
    baseUrl = 'http://localhost:7474/api/session';
    spotifyEndpoint = `${this.baseUrl}/spotify`;
    twitchcEndpoint = `${this.baseUrl}/twitch`;
    accessKeysEndpoint = `${this.baseUrl}/access-keys`;

    constructor(private http: HttpClient) {}

    getOauthStatus() {
        return this.http.get(this.accessKeysEndpoint, {
            withCredentials: true,
        });
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

    setTwitchAccessKey(twitchCode: string) {
        return this.http.post(
            this.twitchcEndpoint,
            { accessKey: twitchCode },
            { withCredentials: true }
        );
    }
}
