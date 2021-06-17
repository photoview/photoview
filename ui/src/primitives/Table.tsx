import styled from 'styled-components'

export const Table = styled.table.attrs({
  className: 'border border-separate rounded' as string,
})`
  border-spacing: 0;

  & td:not(:last-child),
  & th:not(:last-child) {
    border-right: 1px solid;
    border-color: inherit;
  }

  & tr:first-child td {
    border-top: 1px solid;
    border-color: inherit;
  }

  & td {
    border-bottom: 1px solid;
    border-color: inherit;
  }
`

export const TableHeader = styled.thead.attrs({
  className: 'text-left',
})``

export const TableBody = styled.tbody.attrs({ className: '' })``

export const TableFooter = styled.tfoot.attrs({ className: '' })``

export const TableRow = styled.tr.attrs({ className: '' })``

export const TableCell = styled.td.attrs({
  className: 'py-2 px-2 align-top',
})``

export const TableHeaderCell = styled.th.attrs({
  className: 'bg-gray-50 py-2 px-2 align-top font-semibold' as string,
})``
