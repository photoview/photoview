import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TextField } from './Input'

test('a password field can be revealed and hidden again', async () => {
  render(<TextField type="password" defaultValue="hunter2" />)

  const input = screen.getByDisplayValue('hunter2')
  expect(input).toHaveAttribute('type', 'password')

  const toggle = screen.getByRole('button', { name: 'Show password' })
  expect(toggle).toHaveAttribute('aria-pressed', 'false')

  await userEvent.click(toggle)

  expect(input).toHaveAttribute('type', 'text')
  const hideToggle = screen.getByRole('button', { name: 'Hide password' })
  expect(hideToggle).toHaveAttribute('aria-pressed', 'true')

  await userEvent.click(hideToggle)
  expect(input).toHaveAttribute('type', 'password')
})

test('a field of any other type gets no reveal button', () => {
  render(<TextField type="text" defaultValue="not a secret" />)

  expect(screen.queryByRole('button', { name: 'Show password' })).toBeNull()
})

test('a field that stops being a password comes back hidden', async () => {
  const { rerender } = render(
    <TextField type="password" defaultValue="hunter2" />
  )

  await userEvent.click(screen.getByRole('button', { name: 'Show password' }))
  expect(screen.getByDisplayValue('hunter2')).toHaveAttribute('type', 'text')

  rerender(<TextField type="text" defaultValue="hunter2" />)
  rerender(<TextField type="password" defaultValue="hunter2" />)

  // The secret must not reappear unasked.
  expect(screen.getByDisplayValue('hunter2')).toHaveAttribute(
    'type',
    'password'
  )
})

test('the reveal button sits next to the action button without replacing it', async () => {
  const action = vi.fn()
  render(<TextField type="password" defaultValue="hunter2" action={action} />)

  expect(screen.getByRole('button', { name: 'Show password' })).toBeVisible()

  await userEvent.click(screen.getByRole('button', { name: 'Submit' }))
  expect(action).toHaveBeenCalled()
})

test('a revealed password can still be hidden while the field is loading', async () => {
  const { rerender } = render(
    <TextField type="password" defaultValue="hunter2" action={vi.fn()} />
  )

  await userEvent.click(screen.getByRole('button', { name: 'Show password' }))
  expect(screen.getByDisplayValue('hunter2')).toHaveAttribute('type', 'text')

  // Submitting puts the field into its loading state. The spinner stands in
  // for the action button, but the password is on screen and must stay
  // hideable - the alternative is a secret the user cannot put away again.
  rerender(
    <TextField
      type="password"
      defaultValue="hunter2"
      action={vi.fn()}
      loading
    />
  )

  expect(screen.getByLabelText('Loading')).toBeVisible()
  await userEvent.click(screen.getByRole('button', { name: 'Hide password' }))
  expect(screen.getByDisplayValue('hunter2')).toHaveAttribute(
    'type',
    'password'
  )
})

test('revealing and hiding the password keeps the focus in the field', async () => {
  render(<TextField type="password" defaultValue="hunter2" />)

  const input = screen.getByDisplayValue('hunter2')
  input.focus()
  expect(input).toHaveFocus()

  // The toggle swallows mousedown so that clicking it does not blur the field
  // mid-typing; without that, focus would move to the button and the user
  // would have to click back into the field to carry on.
  await userEvent.click(screen.getByRole('button', { name: 'Show password' }))
  expect(input).toHaveFocus()

  await userEvent.click(screen.getByRole('button', { name: 'Hide password' }))
  expect(input).toHaveFocus()
})

test('a revealed password can still be hidden while the field is disabled', async () => {
  const { rerender } = render(
    <TextField type="password" defaultValue="hunter2" />
  )

  await userEvent.click(screen.getByRole('button', { name: 'Show password' }))

  // PasswordProtectedShare disables the field while its request runs. The
  // password is on screen at that moment and must not get stuck there.
  rerender(<TextField type="password" defaultValue="hunter2" disabled />)

  await userEvent.click(screen.getByRole('button', { name: 'Hide password' }))
  expect(screen.getByDisplayValue('hunter2')).toHaveAttribute(
    'type',
    'password'
  )

  // A disabled field that is not revealed cannot be revealed, though.
  expect(screen.getByRole('button', { name: 'Show password' })).toBeDisabled()
})

test('a password field keeps spellcheck and autocorrect away from its value', async () => {
  render(<TextField type="password" defaultValue="hunter2" spellCheck />)

  const input = screen.getByDisplayValue('hunter2')
  await userEvent.click(screen.getByRole('button', { name: 'Show password' }))

  // Revealed, the value is a text field's contents - which enhanced
  // spellcheck in Chrome and Edge sends off to be checked. Even a caller
  // asking for spellcheck does not get it on a password field.
  expect(input).toHaveAttribute('type', 'text')
  expect(input).toHaveAttribute('spellcheck', 'false')
  expect(input).toHaveAttribute('autocorrect', 'off')
  expect(input).toHaveAttribute('autocapitalize', 'off')
})

test('submitting the form masks a revealed password before the browser sees it', async () => {
  let typeAtSubmit: string | null = null

  render(
    <form
      onSubmit={e => {
        e.preventDefault()
        typeAtSubmit = screen.getByDisplayValue('hunter2').getAttribute('type')
      }}
    >
      <TextField type="password" defaultValue="hunter2" />
      <button type="submit">Log in</button>
    </form>
  )

  await userEvent.click(screen.getByRole('button', { name: 'Show password' }))
  expect(screen.getByDisplayValue('hunter2')).toHaveAttribute('type', 'text')

  // Browsers may keep a submitted text field's value as an autocomplete
  // suggestion, so the field has to be a password field again by the time
  // the submission is processed - not after the next render.
  await userEvent.click(screen.getByRole('button', { name: 'Log in' }))

  expect(typeAtSubmit).toBe('password')
  expect(screen.getByDisplayValue('hunter2')).toHaveAttribute(
    'type',
    'password'
  )
})
