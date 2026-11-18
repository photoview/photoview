import styled from 'styled-components'

const Table = styled.table.attrs({ className: 'table-fixed w-full' })``

const Head = styled.thead.attrs({
  className: 'bg-[#f9f9fb] dark:bg-[#2B3037]',
})``

const HeadRow = styled.tr.attrs({
  className:
    'text-left uppercase text-xs border-gray-100 dark:border-dark-border2 border-b border-t',
})``

// A row is not a button: its actions are the buttons inside it, so a screen
// reader announces each of them and a tap on one cell cannot trigger another.
const Row = styled.tr.attrs({
  className: 'border-gray-100 dark:border-dark-border2 border-b',
})``

// Fills its cell, so the row's leading column stays as large a target as the
// whole row used to be.
const RowButton = styled.button.attrs({
  type: 'button',
  className:
    'w-full text-left pl-4 pr-1 py-2 break-words hover:bg-gray-50 focus:outline-none focus-visible:bg-gray-50 dark:hover:bg-[#3c4759] dark:focus-visible:bg-[#3c4759]',
})``

export default {
  Table,
  Head,
  HeadRow,
  Row,
  RowButton,
}
