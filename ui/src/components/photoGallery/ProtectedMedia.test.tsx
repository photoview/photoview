import { render, screen } from "@testing-library/react"
import { ProtectedImage } from "./ProtectedMedia"

const mockPublicUrl = 'http://localhost:9876';
vi.mock('../../helpers/utils', () => ({getPublicUrl: () => mockPublicUrl}));

describe('ProtectedImage', () => {
    test.each([
        {label: 'relative', src: '/image.jpg', expected: `${mockPublicUrl}/image.jpg`},
        {label: 'absolute', src: 'http://localhost:4040/image.jpg', expected: 'http://localhost:4040/image.jpg'}
    ])('loads image correctly given $label url', ({src, expected}) => {

        render(<ProtectedImage alt={'alt_text'} src={src}/>);

        const image = screen.getByAltText('alt_text');

        expect(image).toHaveAttribute('src', expected);
    })
})
