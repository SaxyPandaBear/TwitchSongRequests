import { TestBed } from '@angular/core/testing';

import { OauthService } from './oauth.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('OauthService', () => {
    let service: OauthService;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
        });
        service = TestBed.inject(OauthService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });
});
