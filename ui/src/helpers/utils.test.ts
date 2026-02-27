import { getProtectedUrl, getPublicUrl } from "./utils";

describe('Url helper', () => {
    const mockOrigin = 'http://localhost:9876';
    const mockBaseUrl = '/base_url/';
    const mockShareToken = 'nm42';
    const mockSharePath = `/share/${mockShareToken}`;

    beforeEach(()=>{        
        import.meta.env.BASE_URL = mockBaseUrl;
        Object.defineProperties(window,
            {
                location: {
                    get() {
                        return { origin: mockOrigin, pathname: mockSharePath };
                    }
                },
                document: {
                    get() {
                        return { cookie: 'auth-token=abc' };
                    }
                }
            }
        );
    });

    test.each([
        {url:'', baseUrl: '', expected: `${mockOrigin}/`},
        {url:'', baseUrl: '/public_path/', expected: `${mockOrigin}/public_path/`},
        {url:'/image.jpg', baseUrl: '/public_path/', expected: `${mockOrigin}/public_path/image.jpg`},
        {url:'/image.jpg', baseUrl: 'http://other_host/', expected: 'http://other_host/image.jpg'},
        {url:'http://other_host2/image.jpg', baseUrl: 'http://other_host/', expected: 'http://other_host2/image.jpg'}
    ])('returns currect public url for base $baseUrl and url $url', ({url, baseUrl, expected}) => {
        import.meta.env.BASE_URL = baseUrl;
        const publicUrl = getPublicUrl(url);
        expect(publicUrl.href).toBe(expected);
    })

    test('returns undefined protected url', () => {
       expect(getProtectedUrl(undefined)).toBeUndefined();
    })

    test('returns protected url without token', () => {
        expect(getProtectedUrl('image.jpg')).toBe(`${mockOrigin}${mockBaseUrl}image.jpg`);
    })

    test('returns protected url with token', () => {
        Object.defineProperty(window, 'document', {
            get() {
                return { cookie: '' };
            }
        });
        expect(getProtectedUrl('image.jpg')).toBe(`${mockOrigin}${mockBaseUrl}image.jpg?token=${mockShareToken}`);
    })
})
