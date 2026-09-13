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

  expect(
    screen.getByDisplayValue('hunter2'),
    'the secret must not reappear unasked'
  ).toHaveAttribute('type', 'password')
})

test('the reveal button sits next to the action button without replacing it', async () => {
  const action = vi.fn()
  render(<TextField type="password" defaultValue="hunter2" action={action} />)

  expect(screen.getByRole('button', { name: 'Show password' })).toBeVisible()

  await userEvent.click(screen.getByRole('button', { name: 'Submit' }))
  expect(action).toHaveBeenCalled()
})
