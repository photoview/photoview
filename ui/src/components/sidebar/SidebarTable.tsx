import styled from 'styled-components'

const Table = styled.table.attrs({ className: 'table-fixed w-full' })``

const Head = styled.thead.attrs({
  className: 'bg-[#f9f9fb] dark:bg-[#2B3037]',
})``

const HeadRow = styled.tr.attrs({
  className:
    'text-left uppercase text-xs border-gray-100 dark:border-dark-border2 border-b border-t',
})``

const Row = styled.tr.attrs({
  className:
    'cursor-pointer border-gray-100 dark:border-dark-border2 border-b hover:bg-gray-50 focus:bg-gray-50 dark:hover:bg-[#3c4759] dark:focus:bg-[#3c4759]',
})``

export default {
  Table,
  Head,
  HeadRow,
  Row,
}
