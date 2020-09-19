import { Component, OnInit } from '@angular/core';
import { NgWizardConfig, THEME, NgWizardService } from 'ng-wizard';
import { ActivatedRoute } from '@angular/router';
import { SpotifyService } from '../../spotify.service';
import { OauthService } from '../../oauth.service';
import { take, finalize } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
@Component({
    selector: 'app-login',
    templateUrl: './login.component.html',
    styleUrls: ['./login.component.scss'],
})
export class LoginComponent implements OnInit {
    constructor(
        private route: ActivatedRoute,
        private ngWizardService: NgWizardService,
        private spotifyService: SpotifyService,
        private oauthService: OauthService
    ) {}
    config: NgWizardConfig;

    ngOnInit(): void {
        const { spotifyClientId, twitchClientId } = environment;
        this.spotifyClientId = spotifyClientId;
        this.twitchClientId = twitchClientId;
        this.twitchRedirectPathCode = `https://id.twitch.tv/oauth2/authorize?client_id=${this.twitchClientId}&redirect_uri=${this.localPath}&response_type=code&scope=channel_read+channel:read:redemptions`;
        this.spotifyRedirectUriCode = `https://accounts.spotify.com/authorize?client_id=${this.spotifyClientId}&redirect_uri=${this.localPath}&response_type=code&scope=${this.spotifyScope}`;

        this.oauthService
            .getOauthStatus()
            .pipe(
                take(1),
                finalize(() => {
                    const code = this.route.snapshot.queryParamMap.get('code');
                    if (code) {
                        this.setCode(code);
                    }
                })
            )
            .subscribe((currentUsersAccessTokens) => {
                const {
                    twitchToken,
                    spotifyToken,
                } = currentUsersAccessTokens as any;
                this.spotifyAccessToken = spotifyToken;
                this.twitchAccessToken = twitchToken;
                if (twitchToken) {
                    this.ngWizardService.next();
                }
                if (spotifyToken) {
                    this.ngWizardService.next();
                }
            });
    }
    stepChanged(event) {}

    setSpotifyAccessToken(token) {
        this.oauthService.setSpotifyAcessKey(token).subscribe((res) => {
            this.spotifyAccessToken = true;
            this.ngWizardService.next();
        });
    }
    setTwitchAccessToken(token) {
        this.oauthService.setTwitchAccessKey(token).subscribe((res) => {
            this.twitchAccessToken = true;
            this.ngWizardService.next();
        });
    }

    getUserDevices() {
        this.spotifyService
            .getPlayer(this.spotifyAccessToken)
            .subscribe(
                ({
                    device: { is_active, name: deviceName },
                    item: { name: songName },
                }) => {
                    if (is_active) {
                        this.currentSong = songName;
                        this.playingOn = deviceName;
                        this.playerIsActive = is_active;
                    }
                }
            );
    }

    setCode(code) {
        if (!this.twitchAccessToken) {
            this.setTwitchAccessToken(code);
        } else if (!this.spotifyAccessToken) {
            this.setSpotifyAccessToken(code);
        }
    }

    isLoading = false;

    twitchAccessToken = undefined;
    spotifyScope = 'user-modify-playback-state%20user-read-playback-state';
    localPath = 'http%3A%2F%2Flocalhost%3A4200';
    spotifyClientId;
    twitchClientId;
    twitchRedirectPathCode;

    spotifyRedirectUriCode;

    spotifyAccessToken = undefined;

    currentSong;
    playingOn;
    playerIsActive;
}
