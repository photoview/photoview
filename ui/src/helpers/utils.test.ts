import { getPublicUrl } from "./utils";

describe('getPublicUrl', () => {
    const mockOrigin = 'http://localhost:9876';
    Object.defineProperty(window, 'location', {
        get() {
            return { origin: mockOrigin };
        }
    });

    test.each([
        {label:'relative', baseUrl: '/my_public/', expected: `${mockOrigin}/my_public/`},
        {label:'absolute', baseUrl: 'http://my_origin', expected: 'http://my_origin/'}
    ])('returns currect url for $label base', ({baseUrl, expected}) => {
        import.meta.env.BASE_URL = baseUrl;
        const url = getPublicUrl();
        expect(url.href).toBe(expected);
    })
})
