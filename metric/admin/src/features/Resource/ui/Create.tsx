import {
  GitLabIcon,
  GithubIcon,
  PlusCircleIcon,
  XCircleIcon,
} from "@/shared/ui/Icons";
import { Modal } from "@/shared/ui/Modal";
import { TextInput } from "@/shared/ui/TextInput";
import { FC, useEffect, useState } from "react";
import { PlatformName } from "@/entities/Platform/types";
import { useAppDispatch, useAppSelector } from "@/app/store";
import { resourceSlice } from "@/entities/Resource";
import { useCreateResourceMutation } from "@/entities/Resource/model/resourceApi";
import SelectProvider from "./SelectProvider";
import { shallowEqual } from "react-redux";

const options = [
  {
    name: PlatformName.GitHub,
    icon: <GithubIcon />,
  },
  {
    name: PlatformName.GitLab,
    icon: <GitLabIcon />,
  },
];

interface CreateProps {
  className?: string;
}

const Create: FC<CreateProps> = ({ className = "" }) => {
  const [open, setOpen] = useState(false);

  const [name, setName] = useState("");
  const [provider, setProvider] = useState<string | null>(null);

  const [create, { data: createData }] = useCreateResourceMutation();

  const platforms = useAppSelector(
    (state) => state.platform.platforms,
    shallowEqual,
  );
  const project = useAppSelector(
    (state) => state.project.project,
    shallowEqual,
  );
  const dispatch = useAppDispatch();

  const handleConfirm = async () => {
    if (name && provider && project) {
      // Check if platform exists
      const platform = platforms.find((platform) => platform.name === provider);
      if (!platform) throw new Error("Couldn't find platform");

      // Make a request
      const result = await create({
        name,
        platform: platform.name,
        project: project.id,
      });

      if (!result || result.hasOwnProperty("error"))
        throw new Error("Couldn't create a resource");

      // Set state to empty in case the user would want to create another resource
      setName("");
      setProvider(null);

      setOpen(false);
    }
  };

  // Update redux state on successful creation
  useEffect(() => {
    if (createData && createData?.id) {
      dispatch(resourceSlice.actions.addResource(createData));
    }
  }, [createData, createData?.id]);

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className={
          "bg-[#effbd7] border border-[#e3efcb] text-black px-8 flex text-xl gap-3 justify-center rounded-lg py-4 " +
          className
        }
      >
        <PlusCircleIcon height={27} width={27} />
        <span className="font-bold">New Resource</span>
      </button>
      <Modal show={open} onClose={() => setOpen(false)}>
        <div className="bg-white relative pt-7 w-1/2 rounded-lg pb-11 px-7">
          <div
            className="cursor-pointer absolute top-5 right-5 h-10 w-10"
            onClick={() => setOpen(false)}
          >
            <XCircleIcon />
          </div>
          <h2 className="text-3xl text-[#A3AED0] text-center font-medium">
            Add Provider
          </h2>
          <div className="pt-12"></div>
          <TextInput
            label="Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter the provider name"
          />
          <div className="pt-10"></div>
          <SelectProvider
            options={options}
            selected={provider}
            onSelect={setProvider}
          />
          <div className="pt-14"></div>
          <button
            onClick={handleConfirm}
            className="py-3 px-14 text-[#551FFF] font-medium bg-[#F3F0FF] rounded-lg"
          >
            Confirm
          </button>
        </div>
      </Modal>
    </>
  );
};

export default Create;
