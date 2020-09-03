import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class SpotifyService {
  endpoint = 'https://api.spotify.com/v1/me';
  devicesEndpoint = 'https://api.spotify.com/v1/me/player/devices';
  playerEndpoint = 'https://api.spotify.com/v1/me/player/';

  constructor(private http: HttpClient) {}

  getContactInfo(accessKey) {
    const headers = new HttpHeaders({ Authorization: `Bearer ${accessKey}` });
    return this.http.get(this.endpoint, { headers });
  }
  getDevices(accessKey) {
    const headers = new HttpHeaders({ Authorization: `Bearer ${accessKey}` });
    return this.http.get(this.devicesEndpoint, { headers });
  }
  getPlayer(accessKey): Observable<{ item: any; device: any }> {
    const headers = new HttpHeaders({ Authorization: `Bearer ${accessKey}` });
    return this.http.get(this.playerEndpoint, { headers }) as any;
  }
}
