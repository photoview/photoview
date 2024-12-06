import React, { createContext, useContext, useState } from 'react'

type ModalParams = {
  mediaId: string
}

type ReDetectFacesContextType = {
  isModalOpen: boolean
  openModal: (params?: ModalParams, onClose?: () => void) => void
  closeModal: () => void
  modalParams?: ModalParams | null
}

const ReDetectFacesContext = createContext<ReDetectFacesContextType>({
  isModalOpen: false,
  openModal: () => {
    console.warn('openModal was called before initialized')
  },
  closeModal: () => {
    console.warn('closeModal was called before initialized')
  },
  modalParams: null,
})

export const ReDetectModalProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [modalParams, setModalParams] = useState<ModalParams | null>(null)
  const [onCloseCallback, setOnCloseCallback] = useState<() => void>(() => {
    console.log('Default onCloseCallback executed')
  })

  const openModal = (params?: ModalParams, onClose?: () => void) => {
    setModalParams(params || null)
    setIsModalOpen(true)
    if (onClose) setOnCloseCallback(() => onClose)
  }

  const closeModal = () => {
    setIsModalOpen(false)
    setModalParams(null)
    onCloseCallback()
  }

  return (
    <ReDetectFacesContext.Provider
      value={{ isModalOpen, openModal, closeModal, modalParams }}
    >
      {children}
    </ReDetectFacesContext.Provider>
  )
}

export const useModal = () => useContext(ReDetectFacesContext)
