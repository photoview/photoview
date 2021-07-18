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

export const TableBody = styled.tbody.attrs({ className: '' as string })``

export const TableFooter = styled.tfoot.attrs({ className: '' as string })``

export const TableRow = styled.tr.attrs({ className: '' as string })``

export const TableCell = styled.td.attrs({
  className: 'py-2 px-2 align-top' as string,
})``

export const TableHeaderCell = styled.th.attrs({
  className: 'bg-gray-50 py-2 px-2 align-top font-semibold' as string,
})``

export const TableScrollWrapper = styled.div.attrs({
  className: 'block overflow-x-auto whitespace-nowrap' as string,
})``
